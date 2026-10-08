import type { ApiClient } from './client';
import type { components } from './schema';

export type DirectoryStoreCard = components['schemas']['DirectoryStoreCard'];
export type DirectorySearch = components['schemas']['DirectorySearch'];
export type MatchReason = components['schemas']['MatchReason'];

export interface DirectoryPage<T> {
	items: T[];
	nextCursor: string | null;
	search: DirectorySearch;
}

export async function listDirectoryStores(
	client: ApiClient,
	query: string,
	cursor: string | null
): Promise<DirectoryPage<DirectoryStoreCard>> {
	const params = new URLSearchParams(query);
	if (cursor !== null) {
		params.set('cursor', cursor);
	}
	const suffix = params.size > 0 ? `?${params.toString()}` : '';
	const page = await client.list<DirectoryStoreCard>(`/api/v1/directory/stores${suffix}`);
	return {
		items: page.items,
		nextCursor: page.hasMore ? page.nextCursor : null,
		search: page.search as DirectorySearch
	};
}
