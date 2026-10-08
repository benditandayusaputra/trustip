import type { ApiClient } from '#lib/api/client.js';
import { getSettlement, type Settlement } from '#lib/api/settlements.js';

export class SettlementRecord {
	settlement = $state<Settlement>();
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(settlementId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.settlement = await getSettlement(this.#client(), settlementId);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}
}
