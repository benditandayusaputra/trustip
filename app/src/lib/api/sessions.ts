import type { ApiClient } from './client';
import type { components } from './schema';
import type { Account } from './users';

export type LoginRequest = components['schemas']['LoginRequest'];

export function logIn(client: ApiClient, request: LoginRequest): Promise<Account> {
	return client.send<Account>('/api/v1/sessions', request);
}
