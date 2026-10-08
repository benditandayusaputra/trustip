import type { ApiClient } from './client';
import type { components } from './schema';

export type DropSuggestion = components['schemas']['DropSuggestion'];

export async function listDropSuggestions(
	client: ApiClient,
	partnershipId: string
): Promise<DropSuggestion[]> {
	const path = `/api/v1/partnerships/${encodeURIComponent(partnershipId)}/drop-suggestions`;
	return (await client.list<DropSuggestion>(path)).items;
}
