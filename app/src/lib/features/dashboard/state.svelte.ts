import type { ApiClient } from '#lib/api/client.js';
import {
	listDirectory,
	type DirectoryConsignorCard,
	type DirectoryPath,
	type DirectoryStoreCard,
	type NeedBoardCard
} from '#lib/api/directory.js';
import { listNeeds } from '#lib/api/needs.js';
import { getDashboard, listInsights, type Dashboard, type Insight } from '#lib/api/insights.js';
import { getSalesReport, type SalesReport } from '#lib/api/reports.js';
import type { DateRange } from '#lib/domain/ranges.js';
import type { Side } from './cards';
import {
	isNewPartner,
	needOpportunity,
	opportunitiesPerKind,
	partnerOpportunity,
	responseOpportunity,
	type Opportunity
} from './opportunities';

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

export class OpportunityFeed {
	items = $state<Opportunity[]>([]);
	status = $state<Status>('loading');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(side: Side): Promise<void> {
		this.status = 'loading';
		try {
			this.items = await (side === 'CONSIGNOR'
				? consignorOpportunities(this.#client())
				: storeOpportunities(this.#client()));
			this.status = 'ready';
		} catch {
			this.status = 'failed';
		}
	}
}

function directoryPage<T>(client: ApiClient, path: DirectoryPath, sort: string) {
	return listDirectory<T>(client, { path, query: `sort=${sort}`, cursor: null });
}

async function consignorOpportunities(client: ApiClient): Promise<Opportunity[]> {
	const [needs, stores] = await Promise.all([
		directoryPage<NeedBoardCard>(client, '/api/v1/directory/needs', 'NEAREST'),
		directoryPage<DirectoryStoreCard>(client, '/api/v1/directory/stores', 'BEST_MATCH')
	]);
	return [
		...needs.items
			.filter((need) => !need.is_own)
			.slice(0, opportunitiesPerKind)
			.map(needOpportunity),
		...stores.items
			.filter(isNewPartner)
			.slice(0, opportunitiesPerKind)
			.map((card) => partnerOpportunity(card, '/cari/toko'))
	];
}

async function storeOpportunities(client: ApiClient): Promise<Opportunity[]> {
	const [consignors, needs] = await Promise.all([
		directoryPage<DirectoryConsignorCard>(client, '/api/v1/directory/consignors', 'BEST_MATCH'),
		listNeeds(client, 'OPEN', null)
	]);
	return [
		...needs.items.filter((need) => need.response_count > 0).map(responseOpportunity),
		...consignors.items
			.filter(isNewPartner)
			.slice(0, opportunitiesPerKind)
			.map((card) => partnerOpportunity(card, '/cari/penitip'))
	];
}
