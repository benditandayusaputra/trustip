import type { ApiClient } from '#lib/api/client.js';
import { listDirectory, type DirectoryPath, type DirectorySearch } from '#lib/api/directory.js';

export class DirectoryList<T> {
	items = $state<T[]>([]);
	search = $state<DirectorySearch | null>(null);
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	nextCursor = $state<string | null>(null);
	busy = $state(false);
	#query = '';
	#generation = 0;
	readonly #client: () => ApiClient;
	readonly #path: DirectoryPath;

	constructor(client: () => ApiClient, path: DirectoryPath) {
		this.#client = client;
		this.#path = path;
	}

	async load(query: string): Promise<void> {
		this.#query = query;
		const generation = ++this.#generation;
		this.status = 'loading';
		try {
			const page = await listDirectory<T>(this.#client(), {
				path: this.#path,
				query: this.#query,
				cursor: null
			});
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
			const page = await listDirectory<T>(this.#client(), {
				path: this.#path,
				query: this.#query,
				cursor: this.nextCursor
			});
			this.items = [...this.items, ...page.items];
			this.nextCursor = page.nextCursor;
		} finally {
			this.busy = false;
		}
	}
}
