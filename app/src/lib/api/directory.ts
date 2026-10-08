import type { ApiClient } from './client';
import type { components } from './schema';

export type DirectoryStoreCard = components['schemas']['DirectoryStoreCard'];
export type DirectoryConsignorCard = components['schemas']['DirectoryConsignorCard'];
export type DirectoryProductCard = components['schemas']['DirectoryProductCard'];
export type DirectorySearch = components['schemas']['DirectorySearch'];
export type MatchReason = components['schemas']['MatchReason'];

export type DirectoryPath =
	'/api/v1/directory/stores' | '/api/v1/directory/consignors' | '/api/v1/directory/products';

export interface DirectoryPage<T> {
	items: T[];
	nextCursor: string | null;
	search: DirectorySearch;
}

export interface DirectoryRequest {
	path: DirectoryPath;
	query: string;
	cursor: string | null;
}

export async function listDirectory<T>(
	client: ApiClient,
	request: DirectoryRequest
): Promise<DirectoryPage<T>> {
	const params = new URLSearchParams(request.query);
	if (request.cursor !== null) {
		params.set('cursor', request.cursor);
	}
	const suffix = params.size > 0 ? `?${params.toString()}` : '';
	const page = await client.list<T>(`${request.path}${suffix}`);
	return {
		items: page.items,
		nextCursor: page.hasMore ? page.nextCursor : null,
		search: page.search as DirectorySearch
	};
}

export type DirectoryContact = components['schemas']['DirectoryContact'];

export function getDirectoryContact(
	client: ApiClient,
	organizationId: string
): Promise<DirectoryContact> {
	return client.get<DirectoryContact>(
		`/api/v1/directory/organizations/${encodeURIComponent(organizationId)}`
	);
}
