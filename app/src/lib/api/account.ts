import type { ApiClient } from './client';
import type { components } from './schema';

export type AccountProfile = components['schemas']['AccountProfile'];
export type AccountUpdateRequest = components['schemas']['AccountUpdateRequest'];
export type PasswordChangeRequest = components['schemas']['PasswordChangeRequest'];

export function getAccount(client: ApiClient): Promise<AccountProfile> {
	return client.get<AccountProfile>('/api/v1/me');
}

export function updateAccount(
	client: ApiClient,
	request: AccountUpdateRequest
): Promise<AccountProfile> {
	return client.update<AccountProfile>('/api/v1/me', request);
}

export function changePassword(client: ApiClient, request: PasswordChangeRequest): Promise<null> {
	return client.send<null>('/api/v1/me/password', request);
}
