import type { ApiClient } from '#lib/api/client.js';
import { confirmCount, getCount, type Count } from '#lib/api/counts.js';

function failureOf(error: unknown): string {
	return error instanceof Error ? error.message : '';
}

export class CountRecord {
	count = $state<Count>();
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	busy = $state(false);
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
			this.failure = failureOf(error);
			this.status = 'failed';
		}
	}

	async confirm(): Promise<string | undefined> {
		const current = this.count;
		if (!current) {
			return undefined;
		}
		this.busy = true;
		try {
			this.count = await confirmCount(this.#client(), current.id);
			return undefined;
		} catch (error) {
			return failureOf(error);
		} finally {
			this.busy = false;
		}
	}
}

export function canConfirm(count: Count, viewerOrganizationId: string): boolean {
	return count.status === 'RECORDED' && count.consignor.id === viewerOrganizationId;
}
