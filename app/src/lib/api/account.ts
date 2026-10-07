import type { ApiClient } from './client';
import type { Account } from './users';

export function getAccount(client: ApiClient): Promise<Account> {
	return client.get<Account>('/api/v1/me');
}
