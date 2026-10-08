import { getAccount } from '#lib/api/account.js';
import type { ApiClient } from '#lib/api/client.js';
import { recordCount } from '#lib/api/counts.js';
import { recordDrop } from '#lib/api/drops.js';
import type { ApiError } from '#lib/api/errors.js';
import {
	errorOf,
	problemOf,
	sendingQueues,
	verdictOf,
	type EntryDraft,
	type OutboxEntry,
	type SendVerdict,
	type SentDocument
} from './outbox';

export interface OutboxStorage {
	put: (entry: OutboxEntry) => Promise<void>;
	get: (id: string) => Promise<OutboxEntry | undefined>;
	listFor: (userId: string) => Promise<OutboxEntry[]>;
	remove: (id: string) => Promise<void>;
}

export interface SyncDeps {
	client: () => ApiClient;
	store: OutboxStorage;
	withLock: <T>(task: () => Promise<T>) => Promise<T>;
	now: () => Date;
	isOnline: () => boolean;
	newKey: () => string;
	announce: () => void;
}

export type FlushResult =
	{ kind: 'done'; sent: number } | { kind: 'retryLater'; attempts: number } | { kind: 'signedOut' };

export type SubmitResult =
	| { kind: 'sent'; document: SentDocument }
	| { kind: 'queued' }
	| { kind: 'rejected'; error: ApiError };

type SendOutcome = 'sent' | SendVerdict;

export class OutboxSync {
	readonly #deps: SyncDeps;

	constructor(deps: SyncDeps) {
		this.#deps = deps;
	}

	async enqueue(draft: EntryDraft): Promise<string> {
		const account = await getAccount(this.#deps.client());
		const id = this.#deps.newKey();
		await this.#save({
			...draft,
			id,
			revision: 0,
			idempotencyKey: id,
			recordedAt: this.#deps.now().toISOString(),
			savedOffline: !this.#deps.isOnline(),
			userId: account.id,
			organizationId: account.membership?.organization_id ?? '',
			status: 'PENDING',
			attempts: 0,
			problem: null,
			document: null
		});
		return id;
	}

	flush(): Promise<FlushResult> {
		return this.#deps.withLock(() => this.#flushAll());
	}

	async resultOf(id: string): Promise<SubmitResult> {
		const entry = await this.#deps.store.get(id);
		if (entry?.status === 'SENT' && entry.document) {
			return { kind: 'sent', document: entry.document };
		}
		if (entry?.status === 'NEEDS_REVIEW' && entry.problem) {
			await this.discard(id);
			return { kind: 'rejected', error: errorOf(entry.problem) };
		}
		return { kind: 'queued' };
	}

	async discard(id: string): Promise<void> {
		await this.#deps.store.remove(id);
		this.#deps.announce();
	}

	async revise(entry: OutboxEntry): Promise<void> {
		const key = this.#deps.newKey();
		await this.#save({
			...entry,
			revision: entry.revision + 1,
			idempotencyKey: key,
			status: 'PENDING',
			problem: null
		});
	}

	async #flushAll(): Promise<FlushResult> {
		let userId: string;
		try {
			userId = (await getAccount(this.#deps.client())).id;
		} catch (error) {
			return stopped(verdictOf(error), 0);
		}
		let sent = 0;
		for (const queue of sendingQueues(await this.#deps.store.listFor(userId))) {
			const outcome = await this.#flushQueue(queue);
			if (outcome.kind !== 'done') {
				return outcome;
			}
			sent += outcome.sent;
		}
		return { kind: 'done', sent };
	}

	async #flushQueue(queue: OutboxEntry[]): Promise<FlushResult> {
		let sent = 0;
		for (const [index, entry] of queue.entries()) {
			const outcome = entry.status === 'NEEDS_REVIEW' ? 'review' : await this.#send(entry);
			if (outcome === 'review') {
				await this.#hold(queue.slice(index + 1));
				return { kind: 'done', sent };
			}
			if (outcome !== 'sent') {
				return stopped(outcome, entry.attempts + 1);
			}
			sent += 1;
		}
		return { kind: 'done', sent };
	}

	async #send(entry: OutboxEntry): Promise<SendOutcome> {
		const sending: OutboxEntry = { ...entry, status: 'SENDING' };
		await this.#save(sending);
		try {
			const document = await this.#post(sending);
			await this.#save({ ...sending, status: 'SENT', problem: null, document });
			return 'sent';
		} catch (error) {
			const verdict = verdictOf(error);
			await this.#save({
				...sending,
				status: verdict === 'review' ? 'NEEDS_REVIEW' : 'PENDING',
				attempts: sending.attempts + 1,
				problem: problemOf(error)
			});
			return verdict;
		}
	}

	async #post(entry: OutboxEntry): Promise<SentDocument> {
		const delivery = {
			client_recorded_at: entry.recordedAt,
			recorded_offline: entry.savedOffline || entry.attempts > 0
		};
		const client = this.#deps.client();
		const document =
			entry.kind === 'DROP'
				? await recordDrop(client, { ...entry.request, ...delivery }, entry.idempotencyKey)
				: await recordCount(client, { ...entry.request, ...delivery }, entry.idempotencyKey);
		return { id: document.id, number: document.number };
	}

	async #hold(entries: OutboxEntry[]): Promise<void> {
		for (const entry of entries.filter((waiting) => waiting.status !== 'HELD')) {
			await this.#save({ ...entry, status: 'HELD' });
		}
	}

	async #save(entry: OutboxEntry): Promise<void> {
		await this.#deps.store.put(entry);
		this.#deps.announce();
	}
}

function stopped(verdict: SendVerdict, attempts: number): FlushResult {
	return verdict === 'signedOut' ? { kind: 'signedOut' } : { kind: 'retryLater', attempts };
}
