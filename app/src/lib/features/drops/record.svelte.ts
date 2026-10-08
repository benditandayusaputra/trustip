import type { ApiClient } from '#lib/api/client.js';
import { getDrop, type Drop } from '#lib/api/drops.js';

export class DropRecord {
	drop = $state<Drop>();
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(dropId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.drop = await getDrop(this.#client(), dropId);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}
}
