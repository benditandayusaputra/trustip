import type { ApiClient } from '#lib/api/client.js';
import { listCategories, type Category } from '#lib/api/organizations.js';
import {
	archiveProduct,
	getProduct,
	listProducts,
	reactivateProduct,
	type Product,
	type ProductSearch
} from '#lib/api/products.js';

export type LoadStatus = 'loading' | 'ready' | 'failed';

const searchDelayMs = 300;

export class ProductCatalog {
	items = $state<Product[]>([]);
	categories = $state<Category[]>([]);
	status = $state<LoadStatus>('loading');
	failure = $state('');
	nextCursor = $state<string | null>(null);
	busy = $state(false);
	search = $state<ProductSearch>({ text: '', categoryId: '', status: 'ACTIVE' });
	readonly #client: () => ApiClient;
	#latest = 0;
	#timer: ReturnType<typeof setTimeout> | undefined;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	load(): Promise<void> {
		return this.#loadFor({ ...this.search });
	}

	searchSoon(): void {
		const search = { ...this.search };
		clearTimeout(this.#timer);
		this.#timer = setTimeout(() => void this.#loadFor(search), searchDelayMs);
	}

	async #loadFor(search: ProductSearch): Promise<void> {
		const ticket = ++this.#latest;
		this.status = 'loading';
		try {
			const page = await listProducts(this.#client(), search, null);
			if (ticket !== this.#latest) {
				return;
			}
			this.items = page.items;
			this.nextCursor = page.hasMore ? page.nextCursor : null;
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	async loadCategories(): Promise<void> {
		this.categories = await listCategories(this.#client()).catch(() => []);
	}

	async loadMore(): Promise<void> {
		if (this.nextCursor === null) {
			return;
		}
		this.busy = true;
		try {
			const page = await listProducts(this.#client(), this.search, this.nextCursor);
			this.items = [...this.items, ...page.items];
			this.nextCursor = page.hasMore ? page.nextCursor : null;
		} finally {
			this.busy = false;
		}
	}
}

export class ProductRecord {
	product = $state<Product>();
	status = $state<LoadStatus>('loading');
	failure = $state('');
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(productId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.product = await getProduct(this.#client(), productId);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	async toggleArchive(): Promise<Product> {
		const current = this.product;
		if (!current) {
			throw new Error('product is not loaded');
		}
		this.busy = true;
		try {
			const change = current.status === 'ACTIVE' ? archiveProduct : reactivateProduct;
			this.product = await change(this.#client(), current.id);
			return this.product;
		} finally {
			this.busy = false;
		}
	}
}
