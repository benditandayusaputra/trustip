import type { ApiClient } from './client';
import type { components } from './schema';

export type Count = components['schemas']['Count'];
export type CountRequest = components['schemas']['CountRequest'];
export type CountSheet = components['schemas']['CountSheet'];
export type CountOccasion = components['schemas']['CountOccasion'];

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
