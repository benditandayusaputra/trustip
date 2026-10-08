import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { getAgreement } from '#lib/api/agreements.js';
import { getDrop, listDrops } from '#lib/api/drops.js';
import { listDropSuggestions, type DropSuggestion } from '#lib/api/insights.js';
import { listPartnerCatalog } from '#lib/api/partnerships.js';
import { addDays } from '#lib/domain/dates.js';
import { businessDateOf } from '#lib/i18n/format.js';
import type { EntryDraft, OutboxEntry, QueuedDropLine, SentDocument } from '#lib/offline/outbox.js';
import type { SubmitResult } from '#lib/offline/sync.js';
import { emptyFilter } from './history.svelte';
import {
	defaultExpiry,
	errorKeyOf,
	filledLines,
	linesOf,
	maxDaysBack,
	refilledLines,
	repeatQuantities,
	revisionDateOf,
	toDropRequest,
	validateDrop,
	type DropErrors,
	type DropLine
} from './schema';

export type SaveOutcome =
	| { kind: 'saved'; document: SentDocument }
	| { kind: 'queued' }
	| { kind: 'invalid' }
	| { kind: 'failed'; message: string };

interface OutboxPort {
	submit: (draft: EntryDraft) => Promise<SubmitResult>;
	resubmit: (entryId: string, draft: EntryDraft) => Promise<SubmitResult>;
	find: (entryId: string) => Promise<OutboxEntry | undefined>;
}

interface ComposerDeps {
	client: () => ApiClient;
	outbox: OutboxPort;
	now: () => Date;
}

export class DropComposer {
	consignorName = $state('');
	dropDate = $state('');
	lines = $state<DropLine[]>([]);
	suggestions = $state<DropSuggestion[]>([]);
	errors = $state<DropErrors>({});
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	busy = $state(false);
	#partnershipId = '';
	#agreementVersion = 0;
	#revising = $state('');
	readonly #deps: ComposerDeps;

	constructor(deps: ComposerDeps) {
		this.#deps = deps;
		this.dropDate = this.today;
	}

	get today(): string {
		return businessDateOf(this.#deps.now());
	}

	get earliestDate(): string {
		return addDays(this.today, -maxDaysBack);
	}

	async load(partnershipId: string): Promise<void> {
		this.#partnershipId = partnershipId;
		this.status = 'loading';
		try {
			const [agreement, catalog, suggestions] = await Promise.all([
				getAgreement(this.#deps.client(), partnershipId, 0),
				listPartnerCatalog(this.#deps.client(), partnershipId),
				listDropSuggestions(this.#deps.client(), partnershipId).catch(() => [])
			]);
			this.suggestions = suggestions;
			this.#agreementVersion = agreement.version;
			this.consignorName = agreement.consignor.name;
			this.lines = linesOf(agreement, catalog, this.dropDate);
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
		if (entry?.kind !== 'DROP' || this.status !== 'ready') {
			return;
		}
		this.#revising = entry.id;
		this.changeDate(revisionDateOf(entry.request.occurred_at ?? null, entry.recordedAt));
		this.lines = refilledLines(this.lines, entry.request.items);
	}

	suggestionFor(productId: string): DropSuggestion | undefined {
		return this.suggestions.find((suggestion) => suggestion.product_id === productId);
	}

	changeDate(next: string): void {
		const previous = this.dropDate;
		this.dropDate = next;
		for (const line of this.lines) {
			if (line.expiryDate === defaultExpiry(previous, line.shelfLifeDays)) {
				line.expiryDate = defaultExpiry(next, line.shelfLifeDays);
			}
		}
	}

	async repeatLast(): Promise<string | undefined> {
		const client = this.#deps.client();
		const page = await listDrops(client, emptyFilter(this.#partnershipId), null);
		const last = page.items.find((summary) => summary.status !== 'CANCELLED');
		if (!last) {
			return undefined;
		}
		const previous = await getDrop(client, last.id);
		this.lines = repeatQuantities(this.lines, previous);
		return previous.number;
	}

	updateLine(productId: string, changes: Partial<DropLine>): void {
		this.lines = this.lines.map((line) =>
			line.productId === productId ? { ...line, ...changes } : line
		);
	}

	async save(): Promise<SaveOutcome> {
		this.errors = validateDrop(this.lines, this.dropDate);
		if (Object.keys(this.errors).length > 0) {
			return { kind: 'invalid' };
		}
		const sent = filledLines(this.lines);
		this.busy = true;
		try {
			return this.#outcomeOf(await this.#send(this.#draft(sent)), sent);
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

	#outcomeOf(result: SubmitResult, sent: DropLine[]): SaveOutcome {
		if (result.kind === 'rejected') {
			return this.#explain(result.error, sent);
		}
		return result.kind === 'sent' ? { kind: 'saved', document: result.document } : result;
	}

	#draft(sent: DropLine[]): EntryDraft {
		return {
			kind: 'DROP',
			partnershipId: this.#partnershipId,
			partnerName: this.consignorName,
			request: this.#request(),
			lines: sent.map(queuedLineOf)
		};
	}

	#request() {
		return toDropRequest({
			partnershipId: this.#partnershipId,
			agreementVersion: this.#agreementVersion,
			dropDate: this.dropDate,
			today: this.today,
			lines: this.lines
		});
	}

	#explain(error: unknown, sent: DropLine[]): SaveOutcome {
		if (error instanceof ApiError && error.code === 'AGREEMENT_CHANGED') {
			void this.load(this.#partnershipId);
		}
		if (!(error instanceof ApiError) || error.fields.length === 0) {
			return { kind: 'failed', message: error instanceof Error ? error.message : '' };
		}
		this.errors = Object.fromEntries(
			error.fields.map((problem) => [errorKeyOf(problem.field, sent), problem.message])
		);
		return { kind: 'invalid' };
	}
}

function queuedLineOf(line: DropLine): QueuedDropLine {
	return {
		productId: line.productId,
		name: line.name,
		unit: line.unit,
		qty: line.qty,
		expiryDate: line.expiryDate,
		terms: { ...line.terms }
	};
}
