import type { ApiClient } from './client';
import type { Page } from './envelope';
import type { components } from './schema';

export type Count = components['schemas']['Count'];
export type CountRequest = components['schemas']['CountRequest'];
export type CountSheet = components['schemas']['CountSheet'];
export type CountOccasion = components['schemas']['CountOccasion'];
export type CountSummary = components['schemas']['CountSummary'];

export interface CountFilter {
	partnershipId: string;
	from: string;
	to: string;
}

const countsPath = '/api/v1/counts';

export function getCountSheet(client: ApiClient, partnershipId: string): Promise<CountSheet> {
	return client.get<CountSheet>(
		`/api/v1/partnerships/${encodeURIComponent(partnershipId)}/count-sheet`
	);
}

export function recordCount(
	client: ApiClient,
	request: CountRequest,
	idempotencyKey: string
): Promise<Count> {
	return client.create<Count>(countsPath, request, idempotencyKey);
}

export function getCount(client: ApiClient, countId: string): Promise<Count> {
	return client.get<Count>(`${countsPath}/${encodeURIComponent(countId)}`);
}

export function confirmCount(client: ApiClient, countId: string): Promise<Count> {
	return client.send<Count>(`${countsPath}/${encodeURIComponent(countId)}/confirm`);
}

export function listCounts(
	client: ApiClient,
	filter: CountFilter,
	cursor: string | null
): Promise<Page<CountSummary>> {
	const query = new URLSearchParams();
	const entries: [string, string][] = [
		['partnership_id', filter.partnershipId],
		['from', filter.from],
		['to', filter.to],
		['cursor', cursor ?? '']
	];
	for (const [name, value] of entries) {
		if (value !== '') {
			query.set(name, value);
		}
	}
	const suffix = query.size > 0 ? `?${query.toString()}` : '';
	return client.list<CountSummary>(`${countsPath}${suffix}`);
}
