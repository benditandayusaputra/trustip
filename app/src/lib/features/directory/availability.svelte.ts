import type { ApiClient } from '#lib/api/client.js';
import { getPublicProduct, type PublicProduct } from '#lib/api/public.js';

export type AvailabilityStatus = 'idle' | 'locating' | 'denied' | 'failed' | 'sorted';

export interface GeoPosition {
	latitude: number;
	longitude: number;
}

export interface AvailabilityDeps {
	client: () => ApiClient;
	locate: () => Promise<GeoPosition>;
}

type Stores = PublicProduct['stores'];

export class NearbyStores {
	status = $state<AvailabilityStatus>('idle');
	#sorted = $state<Stores | null>(null);
	readonly #initial: () => Stores;
	readonly #deps: AvailabilityDeps;

	constructor(initial: () => Stores, deps: AvailabilityDeps) {
		this.#initial = initial;
		this.#deps = deps;
	}

	get stores(): Stores {
		return this.#sorted ?? this.#initial();
	}

	async sortFrom(slug: string, sku: string): Promise<void> {
		this.status = 'locating';
		let position: GeoPosition;
		try {
			position = await this.#deps.locate();
		} catch {
			this.status = 'denied';
			return;
		}
		try {
			const product = await getPublicProduct(this.#deps.client(), { slug, sku, near: position });
			this.#sorted = product.stores;
			this.status = 'sorted';
		} catch {
			this.status = 'failed';
		}
	}
}
