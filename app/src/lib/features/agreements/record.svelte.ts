import type { ApiClient } from '#lib/api/client.js';
import { getAgreement, type Agreement } from '#lib/api/agreements.js';

export type LoadStatus = 'loading' | 'ready' | 'failed';

export class AgreementRecord {
	agreement = $state<Agreement>();
	status = $state<LoadStatus>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(partnershipId: string, version = 0): Promise<void> {
		this.status = 'loading';
		try {
			this.agreement = await getAgreement(this.#client(), partnershipId, version);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}
}
