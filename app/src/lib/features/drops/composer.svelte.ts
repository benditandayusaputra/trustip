import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { getAgreement } from '#lib/api/agreements.js';
import { recordDrop, type Drop } from '#lib/api/drops.js';
import { listPartnerCatalog } from '#lib/api/partnerships.js';
import { addDays } from '#lib/domain/dates.js';
import { businessDateOf } from '#lib/i18n/format.js';
import {
	defaultExpiry,
	errorKeyOf,
	filledLines,
	linesOf,
	maxDaysBack,
	toDropRequest,
	validateDrop,
	type DropErrors,
	type DropLine
} from './schema';

export type SaveOutcome =
	{ kind: 'saved'; drop: Drop } | { kind: 'invalid' } | { kind: 'failed'; message: string };

interface ComposerDeps {
	client: () => ApiClient;
	newKey: () => string;
	now: () => Date;
}

export class DropComposer {
	consignorName = $state('');
	dropDate = $state('');
	lines = $state<DropLine[]>([]);
	errors = $state<DropErrors>({});
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	busy = $state(false);
	#partnershipId = '';
	#agreementVersion = 0;
	#key: string;
	readonly #deps: ComposerDeps;

	constructor(deps: ComposerDeps) {
		this.#deps = deps;
		this.#key = deps.newKey();
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
			const [agreement, catalog] = await Promise.all([
				getAgreement(this.#deps.client(), partnershipId, 0),
				listPartnerCatalog(this.#deps.client(), partnershipId)
			]);
			this.#agreementVersion = agreement.version;
			this.consignorName = agreement.consignor.name;
			this.lines = linesOf(agreement, catalog, this.dropDate);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
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

	async save(): Promise<SaveOutcome> {
		this.errors = validateDrop(this.lines, this.dropDate);
		if (Object.keys(this.errors).length > 0) {
			return { kind: 'invalid' };
		}
		const sent = filledLines(this.lines);
		this.busy = true;
		try {
			const drop = await recordDrop(this.#deps.client(), this.#request(), this.#key);
			this.#key = this.#deps.newKey();
			return { kind: 'saved', drop };
		} catch (error) {
			return this.#explain(error, sent);
		} finally {
			this.busy = false;
		}
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
			this.#key = this.#deps.newKey();
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
