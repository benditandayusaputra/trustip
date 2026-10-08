import type { ApiClient } from './client';
import type { Page } from './envelope';
import type { components } from './schema';

export type Settlement = components['schemas']['Settlement'];
export type SettlementSummary = components['schemas']['SettlementSummary'];
type SettlementAge = components['schemas']['SettlementAge'];
export type RunningSales = components['schemas']['RunningSales'];

export interface SettlementFilter {
	partnershipId: string;
	status: '' | 'ISSUED' | 'PAID';
	age: '' | SettlementAge;
}

const settlementsPath = '/api/v1/settlements';

export function getSettlement(client: ApiClient, settlementId: string): Promise<Settlement> {
	return client.get<Settlement>(`${settlementsPath}/${encodeURIComponent(settlementId)}`);
}

export function getRunningSales(client: ApiClient, partnershipId: string): Promise<RunningSales> {
	return client.get<RunningSales>(
		`/api/v1/partnerships/${encodeURIComponent(partnershipId)}/running-sales`
	);
}

export function listSettlements(
	client: ApiClient,
	filter: SettlementFilter,
	cursor: string | null
): Promise<Page<SettlementSummary>> {
	const query = new URLSearchParams();
	const entries: [string, string][] = [
		['partnership_id', filter.partnershipId],
		['status', filter.status],
		['age', filter.age],
		['cursor', cursor ?? '']
	];
	for (const [name, value] of entries) {
		if (value !== '') {
			query.set(name, value);
		}
	}
	const suffix = query.size > 0 ? `?${query.toString()}` : '';
	return client.list<SettlementSummary>(`${settlementsPath}${suffix}`);
}
