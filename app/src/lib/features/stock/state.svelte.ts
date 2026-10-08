import type { ApiClient } from '#lib/api/client.js';
import { listDropSuggestions, type DropSuggestion } from '#lib/api/insights.js';
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
	suggestions = $state<DropSuggestion[]>([]);
	status = $state<LoadStatus>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(partnershipId: string): Promise<void> {
		this.status = 'loading';
		try {
			const [products, suggestions] = await Promise.all([
				getPartnershipStock(this.#client(), partnershipId),
				listDropSuggestions(this.#client(), partnershipId).catch(() => [])
			]);
			this.products = products;
			this.suggestions = suggestions;
			this.status = 'ready';
		} catch (error) {
			this.failure = failureOf(error);
			this.status = 'failed';
		}
	}
}

export function suggestedQtyOf(
	suggestions: DropSuggestion[],
	productId: string
): number | undefined {
	const found = suggestions.find((suggestion) => suggestion.product_id === productId);
	return found?.has_enough_data === true ? found.suggested_qty : undefined;
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
