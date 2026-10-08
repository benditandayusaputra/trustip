import type { ApiClient } from '#lib/api/client.js';
import { getCount, type Count } from '#lib/api/counts.js';

export class CountRecord {
	count = $state<Count>();
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(countId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.count = await getCount(this.#client(), countId);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}
}
