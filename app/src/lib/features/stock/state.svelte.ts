import type { ApiClient } from '#lib/api/client.js';
import {
	getPartnershipStock,
	getStockCard,
	listStock,
	type CardKey,
	type ProductStock,
	type StockCard,
	type StockHolding
} from '#lib/api/stock.js';

type LoadStatus = 'loading' | 'ready' | 'failed';

function failureOf(error: unknown): string {
	return error instanceof Error ? error.message : '';
}

export class StockHoldings {
	items = $state<StockHolding[]>([]);
	status = $state<LoadStatus>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(): Promise<void> {
		this.status = 'loading';
		try {
			this.items = await listStock(this.#client());
			this.status = 'ready';
		} catch (error) {
			this.failure = failureOf(error);
			this.status = 'failed';
		}
	}
}

export class PartnershipStock {
	products = $state<ProductStock[]>([]);
	status = $state<LoadStatus>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(partnershipId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.products = await getPartnershipStock(this.#client(), partnershipId);
			this.status = 'ready';
		} catch (error) {
			this.failure = failureOf(error);
			this.status = 'failed';
		}
	}
}

export class StockCardRecord {
	card = $state<StockCard>();
	status = $state<LoadStatus>('loading');
	failure = $state('');
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(key: CardKey): Promise<void> {
		this.status = 'loading';
		try {
			this.card = await getStockCard(this.#client(), key, null);
			this.status = 'ready';
		} catch (error) {
			this.failure = failureOf(error);
			this.status = 'failed';
		}
	}

	async loadMore(key: CardKey): Promise<void> {
		const current = this.card;
		if (!current?.next_cursor) {
			return;
		}
		this.busy = true;
		try {
			const next = await getStockCard(this.#client(), key, current.next_cursor);
			this.card = { ...next, entries: [...current.entries, ...next.entries] };
		} finally {
			this.busy = false;
		}
	}
}
