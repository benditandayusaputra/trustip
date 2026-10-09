import type { ApiClient } from './client';
import type { components } from './schema';

export type PublicStore = components['schemas']['PublicStore'];
export type PublicConsignor = components['schemas']['PublicConsignor'];

export function getPublicStore(client: ApiClient, slug: string): Promise<PublicStore> {
	return client.get<PublicStore>(`/api/v1/public/stores/${encodeURIComponent(slug)}`);
}

export function getPublicConsignor(client: ApiClient, slug: string): Promise<PublicConsignor> {
	return client.get<PublicConsignor>(`/api/v1/public/consignors/${encodeURIComponent(slug)}`);
}

export type PublicProduct = components['schemas']['PublicProduct'];
export type PublicNeed = components['schemas']['PublicNeed'];

export function getPublicNeed(client: ApiClient, needId: string): Promise<PublicNeed> {
	return client.get<PublicNeed>(`/api/v1/public/needs/${encodeURIComponent(needId)}`);
}

export interface ProductAddress {
	slug: string;
	sku: string;
	near?: { latitude: number; longitude: number };
}

export function getPublicProduct(
	client: ApiClient,
	address: ProductAddress
): Promise<PublicProduct> {
	const path = `/api/v1/public/products/${encodeURIComponent(address.slug)}/${encodeURIComponent(address.sku)}`;
	if (!address.near) {
		return client.get<PublicProduct>(path);
	}
	const near = new URLSearchParams({
		lat: String(address.near.latitude),
		lng: String(address.near.longitude)
	});
	return client.get<PublicProduct>(`${path}?${near.toString()}`);
}
