import type { ApiClient } from '#lib/api/client.js';
import {
	getDailyRecap,
	getSettlementReport,
	type DailyRecap,
	type SettlementReport
} from '#lib/api/reports.js';
import { rangeOf, type DateRange, type RangePreset } from '#lib/domain/ranges.js';
import { businessDateOf } from '#lib/i18n/format.js';

type LoadStatus = 'loading' | 'ready' | 'failed';

function failureOf(error: unknown): string {
	return error instanceof Error ? error.message : '';
}

export class DailyRecapState {
	date = $state('');
	recap = $state<DailyRecap>();
	status = $state<LoadStatus>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient, now: () => Date) {
		this.#client = client;
		this.date = businessDateOf(now());
	}

	async load(date: string): Promise<void> {
		this.status = 'loading';
		try {
			this.recap = await getDailyRecap(this.#client(), date);
			this.status = 'ready';
		} catch (error) {
			this.failure = failureOf(error);
			this.status = 'failed';
		}
	}
}

export class SettlementReportState {
	preset = $state<RangePreset>('THIS_MONTH');
	custom = $state<DateRange>({ from: '', to: '' });
	report = $state<SettlementReport>();
	status = $state<LoadStatus>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;
	readonly #now: () => Date;

	constructor(client: () => ApiClient, now: () => Date) {
		this.#client = client;
		this.#now = now;
		const today = businessDateOf(now());
		this.custom = { from: today, to: today };
	}

	get range(): DateRange {
		return rangeOf(this.preset, businessDateOf(this.#now()), this.custom);
	}

	async load(range: DateRange): Promise<void> {
		this.status = 'loading';
		try {
			this.report = await getSettlementReport(this.#client(), range);
			this.status = 'ready';
		} catch (error) {
			this.failure = failureOf(error);
			this.status = 'failed';
		}
	}
}
