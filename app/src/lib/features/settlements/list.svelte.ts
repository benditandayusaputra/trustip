import type { ApiClient } from '#lib/api/client.js';
import {
	getRunningSales,
	listSettlements,
	type RunningSales,
	type SettlementFilter,
	type SettlementSummary
} from '#lib/api/settlements.js';

export function emptyFilter(partnershipId = ''): SettlementFilter {
	return { partnershipId, status: '', age: '' };
}

export class SettlementList {
	items = $state<SettlementSummary[]>([]);
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	nextCursor = $state<string | null>(null);
	busy = $state(false);
	#filter: SettlementFilter = emptyFilter();
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(filter: SettlementFilter): Promise<void> {
		this.#filter = { ...filter };
		this.status = 'loading';
		try {
			const page = await listSettlements(this.#client(), this.#filter, null);
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
			const page = await listSettlements(this.#client(), this.#filter, this.nextCursor);
			this.items = [...this.items, ...page.items];
			this.nextCursor = page.hasMore ? page.nextCursor : null;
		} finally {
			this.busy = false;
		}
	}
}

export class RunningRecord {
	running = $state<RunningSales>();
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(partnershipId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.running = await getRunningSales(this.#client(), partnershipId);
			this.status = 'ready';
		} catch {
			this.status = 'failed';
		}
	}
}
