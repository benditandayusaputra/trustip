import type { ApiClient } from '#lib/api/client.js';
import { listCounts, type CountFilter, type CountSummary } from '#lib/api/counts.js';
import { listPartnerships, type PartnershipSummary } from '#lib/api/partnerships.js';

export function emptyFilter(partnershipId = ''): CountFilter {
	return { partnershipId, from: '', to: '' };
}

export class CountHistory {
	items = $state<CountSummary[]>([]);
	partners = $state<PartnershipSummary[]>([]);
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	nextCursor = $state<string | null>(null);
	busy = $state(false);
	#filter: CountFilter = emptyFilter();
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async loadPartners(): Promise<void> {
		const page = await listPartnerships(this.#client(), '', null).catch(() => undefined);
		this.partners = page?.items ?? [];
	}

	async load(filter: CountFilter): Promise<void> {
		this.#filter = { ...filter };
		this.status = 'loading';
		try {
			const page = await listCounts(this.#client(), this.#filter, null);
			this.items = page.items;
			this.nextCursor = page.hasMore ? page.nextCursor : null;
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	async loadMore(): Promise<void> {
		if (this.nextCursor === null) {
			return;
		}
		this.busy = true;
		try {
			const page = await listCounts(this.#client(), this.#filter, this.nextCursor);
			this.items = [...this.items, ...page.items];
			this.nextCursor = page.hasMore ? page.nextCursor : null;
		} finally {
			this.busy = false;
		}
	}
}
