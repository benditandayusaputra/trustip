import type { ApiClient } from '#lib/api/client.js';
import { getDashboard, listInsights, type Dashboard, type Insight } from '#lib/api/insights.js';
import { getSalesReport, type SalesReport } from '#lib/api/reports.js';
import type { DateRange } from '#lib/domain/ranges.js';

type Status = 'loading' | 'ready' | 'failed';

export class DashboardState {
	dashboard = $state<Dashboard>();
	insights = $state<Insight[]>([]);
	status = $state<Status>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(): Promise<void> {
		this.status = 'loading';
		try {
			const [dashboard, insights] = await Promise.all([
				getDashboard(this.#client()),
				listInsights(this.#client())
			]);
			this.dashboard = dashboard;
			this.insights = insights;
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}
}

export class SalesOverview {
	byPartner = $state<SalesReport>();
	byProduct = $state<SalesReport>();
	status = $state<Status>('loading');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(range: DateRange): Promise<void> {
		this.status = 'loading';
		try {
			const [byPartner, byProduct] = await Promise.all([
				getSalesReport(this.#client(), { ...range, groupBy: 'PARTNER' }),
				getSalesReport(this.#client(), { ...range, groupBy: 'PRODUCT' })
			]);
			this.byPartner = byPartner;
			this.byProduct = byProduct;
			this.status = 'ready';
		} catch {
			this.status = 'failed';
		}
	}
}
