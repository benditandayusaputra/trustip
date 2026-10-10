import type { Tone } from '#lib/components/ui/tone.js';
import { offline as id } from '#lib/i18n/offline.js';
import { isOld, isWaiting, type OutboxEntry } from '#lib/offline/outbox.js';
import type { FlushResult } from '#lib/offline/sync.js';

interface QueueDeps {
	list: () => Promise<OutboxEntry[]>;
	send: () => Promise<FlushResult>;
	discard: (entryId: string) => Promise<FlushResult>;
}

const statusTones: Record<OutboxEntry['status'], Tone> = {
	PENDING: 'pending',
	SENDING: 'info',
	SENT: 'success',
	HELD: 'neutral',
	NEEDS_REVIEW: 'danger'
};

export function statusToneOf(entry: OutboxEntry): Tone {
	return statusTones[entry.status];
}

function sendMessageOf(result: FlushResult): string {
	const text = id.outbox;
	if (result.kind === 'done') {
		return text.sent(result.sent);
	}
	return result.kind === 'signedOut' ? text.signedOut : text.retryLater;
}

function byRecording(left: OutboxEntry, right: OutboxEntry): number {
	return left.recordedAt.localeCompare(right.recordedAt);
}

export class OutboxQueue {
	entries = $state<OutboxEntry[]>([]);
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	sending = $state(false);
	readonly #deps: QueueDeps;

	constructor(deps: QueueDeps) {
		this.#deps = deps;
	}

	oldCount(now: Date): number {
		return this.entries.filter((entry) => isOld(entry, now)).length;
	}

	get waitingCount(): number {
		return this.entries.filter(isWaiting).length;
	}

	async load(): Promise<void> {
		try {
			this.entries = (await this.#deps.list()).sort(byRecording);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	async sendNow(): Promise<string> {
		this.sending = true;
		try {
			return sendMessageOf(await this.#deps.send());
		} finally {
			this.sending = false;
			await this.load();
		}
	}

	async discard(entryId: string): Promise<void> {
		await this.#deps.discard(entryId);
		await this.load();
	}
}
