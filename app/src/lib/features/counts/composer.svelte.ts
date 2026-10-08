import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { getCountSheet, type CountOccasion, type CountSheet } from '#lib/api/counts.js';
import { businessDateOf } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';
import type { EntryDraft, OutboxEntry, SentDocument } from '#lib/offline/outbox.js';
import type { SubmitResult } from '#lib/offline/sync.js';
import { withQueuedDrops } from './queued';
import {
	allSold,
	earliestCountDate,
	errorKeyOf,
	linesOf,
	refilledCountLines,
	returnAllRemaining,
	revisionDateOf,
	withExpiredReturn,
	type ExpiredReturn,
	toCountRequest,
	validateCount,
	withChanges,
	type CountErrors,
	type CountLineValues,
	type InputMode
} from './schema';

const reloadCodes = new Set(['STOCK_CHANGED', 'AGREEMENT_CHANGED']);

export type SaveOutcome =
	| { kind: 'saved'; document: SentDocument }
	| { kind: 'queued' }
	| { kind: 'invalid' }
	| { kind: 'failed'; message: string };

interface OutboxPort {
	submit: (draft: EntryDraft) => Promise<SubmitResult>;
	resubmit: (entryId: string, draft: EntryDraft) => Promise<SubmitResult>;
	find: (entryId: string) => Promise<OutboxEntry | undefined>;
	waiting: () => Promise<OutboxEntry[]>;
}

interface ComposerDeps {
	client: () => ApiClient;
	outbox: OutboxPort;
	now: () => Date;
}

type LineChanges = Parameters<typeof withChanges>[1];

export class CountComposer {
	sheet = $state<CountSheet>();
	lines = $state<CountLineValues[]>([]);
	countedOn = $state('');
	occasion = $state<CountOccasion>('STORE_CLOSING');
	note = $state('');
	mode = $state<InputMode>('remaining');
	errors = $state<CountErrors>({});
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	busy = $state(false);
	#partnershipId = '';
	#revising = $state('');
	readonly #deps: ComposerDeps;

	constructor(deps: ComposerDeps) {
		this.#deps = deps;
		this.countedOn = this.today;
	}

	get today(): string {
		return businessDateOf(this.#deps.now());
	}

	get earliestDate(): string {
		return earliestCountDate(this.today, this.sheet?.last_document_at ?? null);
	}

	async load(partnershipId: string): Promise<void> {
		this.#partnershipId = partnershipId;
		this.status = 'loading';
		try {
			const [sheet, waiting] = await Promise.all([
				getCountSheet(this.#deps.client(), partnershipId),
				this.#deps.outbox.waiting()
			]);
			this.sheet = withQueuedDrops(sheet, waiting);
			this.lines = linesOf(this.sheet);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	get isRevising(): boolean {
		return this.#revising !== '';
	}

	async loadRevision(partnershipId: string, entryId: string): Promise<void> {
		await this.load(partnershipId);
		const entry = await this.#deps.outbox.find(entryId);
		if (entry?.kind !== 'COUNT' || this.status !== 'ready') {
			return;
		}
		this.#revising = entry.id;
		this.occasion = entry.request.occasion;
		this.note = entry.request.note ?? '';
		this.countedOn = revisionDateOf(entry.request.counted_at ?? null, entry.recordedAt);
		this.lines = refilledCountLines(this.lines, entry.request.lines);
	}

	updateLine(productId: string, changes: LineChanges): void {
		this.lines = this.lines.map((line) =>
			line.productId === productId ? withChanges(line, changes) : line
		);
	}

	markAllSold(): void {
		this.lines = allSold(this.lines);
	}

	returnExpired(expired: ExpiredReturn): void {
		this.lines = withExpiredReturn(this.lines, expired);
	}

	returnAllRemaining(): void {
		this.lines = returnAllRemaining(this.lines);
	}

	async save(): Promise<SaveOutcome> {
		this.errors = validateCount(this.lines);
		if (Object.keys(this.errors).length > 0 || !this.sheet) {
			return { kind: 'invalid' };
		}
		const draft: EntryDraft = {
			kind: 'COUNT',
			partnershipId: this.#partnershipId,
			partnerName: this.sheet.consignor.name,
			request: this.#request(this.sheet)
		};
		this.busy = true;
		try {
			return this.#outcomeOf(await this.#send(draft));
		} catch (error) {
			return { kind: 'failed', message: error instanceof Error ? error.message : '' };
		} finally {
			this.busy = false;
		}
	}

	#send(draft: EntryDraft): Promise<SubmitResult> {
		const outbox = this.#deps.outbox;
		return this.isRevising ? outbox.resubmit(this.#revising, draft) : outbox.submit(draft);
	}

	#outcomeOf(result: SubmitResult): SaveOutcome {
		if (result.kind === 'rejected') {
			return this.#explain(result.error);
		}
		return result.kind === 'sent' ? { kind: 'saved', document: result.document } : result;
	}

	#request(sheet: CountSheet) {
		return toCountRequest({
			partnershipId: this.#partnershipId,
			agreementVersion: sheet.agreement_version,
			countedOn: this.countedOn,
			today: this.today,
			occasion: this.occasion,
			note: this.note,
			lines: this.lines
		});
	}

	#explain(error: unknown): SaveOutcome {
		if (error instanceof ApiError && reloadCodes.has(error.code)) {
			void this.load(this.#partnershipId);
			const message = error.code === 'STOCK_CHANGED' ? id.counts.stockChanged : error.message;
			return { kind: 'failed', message };
		}
		if (!(error instanceof ApiError) || error.fields.length === 0) {
			return { kind: 'failed', message: error instanceof Error ? error.message : '' };
		}
		const lines = this.lines;
		this.errors = Object.fromEntries(
			error.fields.map((problem) => [errorKeyOf(problem.field, lines), problem.message])
		);
		return { kind: 'invalid' };
	}
}
