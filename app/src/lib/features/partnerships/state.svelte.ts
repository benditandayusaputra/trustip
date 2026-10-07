import type { ApiClient } from '#lib/api/client.js';
import {
	getPartnership,
	listPartnerCatalog,
	listPartnerships,
	type PartnerCatalogItem,
	type Partnership,
	type PartnershipSummary
} from '#lib/api/partnerships.js';

export type LoadStatus = 'loading' | 'ready' | 'failed';
export type PartnerFilter = 'ACTIVE' | 'ENDED' | 'ALL';

function failureOf(error: unknown): string {
	return error instanceof Error ? error.message : '';
}

export class PartnerList {
	items = $state<PartnershipSummary[]>([]);
	status = $state<LoadStatus>('loading');
	failure = $state('');
	nextCursor = $state<string | null>(null);
	busy = $state(false);
	#filter: PartnerFilter = 'ACTIVE';
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(filter: PartnerFilter): Promise<void> {
		this.#filter = filter;
		this.status = 'loading';
		try {
			const page = await listPartnerships(this.#client(), this.#status(), null);
			this.items = page.items;
			this.nextCursor = page.hasMore ? page.nextCursor : null;
			this.status = 'ready';
		} catch (error) {
			this.failure = failureOf(error);
			this.status = 'failed';
		}
	}

	async loadMore(): Promise<void> {
		if (this.nextCursor === null) {
			return;
		}
		this.busy = true;
		try {
			const page = await listPartnerships(this.#client(), this.#status(), this.nextCursor);
			this.items = [...this.items, ...page.items];
			this.nextCursor = page.hasMore ? page.nextCursor : null;
		} finally {
			this.busy = false;
		}
	}

	#status(): PartnershipSummary['status'] | '' {
		return this.#filter === 'ALL' ? '' : this.#filter;
	}
}

export class PartnerRecord {
	partnership = $state<Partnership>();
	catalog = $state<PartnerCatalogItem[]>([]);
	status = $state<LoadStatus>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(partnershipId: string): Promise<void> {
		this.status = 'loading';
		try {
			const [partnership, catalog] = await Promise.all([
				getPartnership(this.#client(), partnershipId),
				listPartnerCatalog(this.#client(), partnershipId)
			]);
			this.partnership = partnership;
			this.catalog = catalog;
			this.status = 'ready';
		} catch (error) {
			this.failure = failureOf(error);
			this.status = 'failed';
		}
	}
}
