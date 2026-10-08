import type { ApiClient } from '#lib/api/client.js';
import {
	listDirectoryStores,
	type DirectorySearch,
	type DirectoryStoreCard
} from '#lib/api/directory.js';

export class StoreDirectory {
	items = $state<DirectoryStoreCard[]>([]);
	search = $state<DirectorySearch | null>(null);
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	nextCursor = $state<string | null>(null);
	busy = $state(false);
	#query = '';
	#generation = 0;
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(query: string): Promise<void> {
		this.#query = query;
		const generation = ++this.#generation;
		this.status = 'loading';
		try {
			const page = await listDirectoryStores(this.#client(), this.#query, null);
			if (generation !== this.#generation) {
				return;
			}
			this.items = page.items;
			this.search = page.search;
			this.nextCursor = page.nextCursor;
			this.status = 'ready';
		} catch (error) {
			if (generation === this.#generation) {
				this.failure = error instanceof Error ? error.message : '';
				this.status = 'failed';
			}
		}
	}

	async loadMore(): Promise<void> {
		if (this.nextCursor === null || this.busy) {
			return;
		}
		this.busy = true;
		try {
			const page = await listDirectoryStores(this.#client(), this.#query, this.nextCursor);
			this.items = [...this.items, ...page.items];
			this.nextCursor = page.nextCursor;
		} finally {
			this.busy = false;
		}
	}
}
