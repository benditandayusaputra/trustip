import type { ApiClient } from '#lib/api/client.js';
import { listPartnerships, type PartnershipSummary } from '#lib/api/partnerships.js';

export class ActivePartners {
	items = $state<PartnershipSummary[]>([]);
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(): Promise<void> {
		this.status = 'loading';
		try {
			this.items = (await listPartnerships(this.#client(), 'ACTIVE', null)).items;
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}
}
