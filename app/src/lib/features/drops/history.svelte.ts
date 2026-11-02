import { untrack } from 'svelte';
import type { ApiClient } from '#lib/api/client.js';
import { listDrops, type DropFilter, type DropSummary } from '#lib/api/drops.js';

export function emptyFilter(partnershipId = ''): DropFilter {
	return { partnershipId, status: '', from: '', to: '' };
}

export class DropHistory {
	items = $state<DropSummary[]>([]);
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	nextCursor = $state<string | null>(null);
	busy = $state(false);
	#filter: DropFilter = emptyFilter();
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(filter: DropFilter): Promise<void> {
		this.#filter = { ...filter };
		if (untrack(() => this.status) !== 'ready') {
			this.status = 'loading';
		}
		try {
			const page = await listDrops(this.#client(), this.#filter, null);
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
			const page = await listDrops(this.#client(), this.#filter, this.nextCursor);
			this.items = [...this.items, ...page.items];
			this.nextCursor = page.hasMore ? page.nextCursor : null;
		} finally {
			this.busy = false;
		}
	}
}
