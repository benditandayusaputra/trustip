import type { ApiClient } from './client';
import type { components } from './schema';
import type { Account } from './users';

export type LoginRequest = components['schemas']['LoginRequest'];

export function logIn(client: ApiClient, request: LoginRequest): Promise<Account> {
	return client.send<Account>('/api/v1/sessions', request);
}

export function logOut(client: ApiClient): Promise<null> {
	return client.send<null>('/api/v1/sessions/logout');
}

export function logOutEverywhere(client: ApiClient): Promise<null> {
	return client.send<null>('/api/v1/sessions/logout-all');
}
