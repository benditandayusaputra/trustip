import type { ApiClient } from './client';
import type { components } from './schema';

export type Drop = components['schemas']['Drop'];
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
