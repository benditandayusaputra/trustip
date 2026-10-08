import type { ApiClient } from './client';
import type { Page } from './envelope';
import type { components } from './schema';

export type Need = components['schemas']['Need'];
export type NeedDetail = components['schemas']['NeedDetail'];
export type NeedRequest = components['schemas']['NeedRequest'];
export type NeedStatus = components['schemas']['NeedStatus'];

const needsPath = '/api/v1/needs';

function needPath(needId: string): string {
	return `${needsPath}/${encodeURIComponent(needId)}`;
}

export function postNeed(client: ApiClient, request: NeedRequest): Promise<Need> {
	return client.send<Need>(needsPath, request);
}

export function getNeed(client: ApiClient, needId: string): Promise<NeedDetail> {
	return client.get<NeedDetail>(needPath(needId));
}

export function listNeeds(
	client: ApiClient,
	status: NeedStatus | '',
	cursor: string | null
): Promise<Page<Need>> {
	const query = new URLSearchParams();
	if (status !== '') {
		query.set('status', status);
	}
	if (cursor !== null) {
		query.set('cursor', cursor);
	}
	const suffix = query.size > 0 ? `?${query.toString()}` : '';
	return client.list<Need>(`${needsPath}${suffix}`);
}

export type NeedAction = 'extend' | 'fulfill' | 'close';

export function changeNeed(client: ApiClient, needId: string, action: NeedAction): Promise<Need> {
	return client.send<Need>(`${needPath(needId)}/${action}`);
}
