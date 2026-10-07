import type { ApiClient } from './client';
import type { components } from './schema';

export type RegisterRequest = components['schemas']['RegisterRequest'];
export type Account = components['schemas']['Account'];

export function registerUser(client: ApiClient, request: RegisterRequest): Promise<Account> {
	return client.send<Account>('/api/v1/users', request);
}
