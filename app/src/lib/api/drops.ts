import type { ApiClient } from './client';
import type { Page } from './envelope';
import type { components } from './schema';

export type Drop = components['schemas']['Drop'];
export type DropSummary = components['schemas']['DropSummary'];
type DropStatus = components['schemas']['DropStatus'];

export interface DropFilter {
	partnershipId: string;
	status: DropStatus | '';
	from: string;
	to: string;
}
export type DropRequest = components['schemas']['DropRequest'];

const dropsPath = '/api/v1/drops';

export function recordDrop(
	client: ApiClient,
	request: DropRequest,
	idempotencyKey: string
): Promise<Drop> {
	return client.create<Drop>(dropsPath, request, idempotencyKey);
}

export function getDrop(client: ApiClient, dropId: string): Promise<Drop> {
	return client.get<Drop>(`${dropsPath}/${encodeURIComponent(dropId)}`);
}

export function confirmDrop(client: ApiClient, dropId: string): Promise<Drop> {
	return client.send<Drop>(`${dropsPath}/${encodeURIComponent(dropId)}/confirm`);
}

export function cancelDrop(client: ApiClient, dropId: string): Promise<Drop> {
	return client.send<Drop>(`${dropsPath}/${encodeURIComponent(dropId)}/cancel`);
}

export function listDrops(
	client: ApiClient,
	filter: DropFilter,
	cursor: string | null
): Promise<Page<DropSummary>> {
	const query = new URLSearchParams();
	const entries: [string, string][] = [
		['partnership_id', filter.partnershipId],
		['status', filter.status],
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
	return client.list<DropSummary>(`${dropsPath}${suffix}`);
}
