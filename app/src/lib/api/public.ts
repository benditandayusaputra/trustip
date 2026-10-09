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
