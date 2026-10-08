import type { ApiClient } from '#lib/api/client.js';
import { getSalesReport, type ReportGrouping, type SalesReport } from '#lib/api/reports.js';
import { rangeOf, type DateRange, type RangePreset } from '#lib/domain/ranges.js';
import { businessDateOf } from '#lib/i18n/format.js';

export class SalesReportView {
	preset = $state<RangePreset>('LAST_7');
	custom = $state<DateRange>({ from: '', to: '' });
	groupBy = $state<ReportGrouping>('PARTNER');
	report = $state<SalesReport>();
	status = $state<'loading' | 'ready' | 'failed'>('loading');
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

	async load(range: DateRange, groupBy: ReportGrouping): Promise<void> {
		this.status = 'loading';
		try {
			this.report = await getSalesReport(this.#client(), { ...range, groupBy });
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}
}
