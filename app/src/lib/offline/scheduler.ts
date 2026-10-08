import { retryDelayMs, type EntryDraft } from './outbox';
import type { FlushResult, OutboxSync, SubmitResult } from './sync';

interface Timers {
	set: (task: () => void, delayMs: number) => number;
	clear: (handle: number) => void;
}

export interface SchedulerDeps {
	sync: OutboxSync;
	afterSent: () => void;
	keepStorage: () => void;
	timers: Timers;
	random: () => number;
}

export class OutboxScheduler {
	readonly #deps: SchedulerDeps;
	#retry: number | undefined;

	constructor(deps: SchedulerDeps) {
		this.#deps = deps;
	}

	async submit(draft: EntryDraft): Promise<SubmitResult> {
		const id = await this.#deps.sync.enqueue(draft);
		await this.run();
		const result = await this.#deps.sync.resultOf(id);
		if (result.kind === 'queued') {
			this.#deps.keepStorage();
		}
		return result;
	}

	async resubmit(entryId: string, draft: EntryDraft): Promise<SubmitResult> {
		await this.#deps.sync.revise(entryId, draft);
		await this.run();
		return this.#deps.sync.resultOf(entryId);
	}

	async run(): Promise<FlushResult> {
		this.stop();
		const result = await this.#deps.sync.flush();
		if (result.kind === 'retryLater') {
			const delay = retryDelayMs(result.attempts, this.#deps.random());
			this.#retry = this.#deps.timers.set(() => void this.run(), delay);
		}
		if (result.kind === 'done' && result.sent > 0) {
			this.#deps.afterSent();
		}
		return result;
	}

	async discard(id: string): Promise<FlushResult> {
		await this.#deps.sync.discard(id);
		return this.run();
	}

	stop(): void {
		if (this.#retry !== undefined) {
			this.#deps.timers.clear(this.#retry);
			this.#retry = undefined;
		}
	}
}
