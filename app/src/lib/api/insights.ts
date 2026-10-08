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

export type Insight = components['schemas']['Insight'];
export type Dashboard = components['schemas']['Dashboard'];

export async function listInsights(client: ApiClient): Promise<Insight[]> {
	return (await client.list<Insight>('/api/v1/insights')).items;
}

export function getDashboard(client: ApiClient): Promise<Dashboard> {
	return client.get<Dashboard>('/api/v1/dashboard');
}
