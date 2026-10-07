import type { ApiClient } from './client';
import type { Page } from './envelope';
import type { components } from './schema';

export type Notification = components['schemas']['Notification'];
type UnreadCount = components['schemas']['UnreadCount'];

export function listNotifications(
	client: ApiClient,
	cursor: string | null
): Promise<Page<Notification>> {
	const query = cursor ? `?${new URLSearchParams({ cursor }).toString()}` : '';
	return client.list<Notification>(`/api/v1/notifications${query}`);
}

export async function countUnread(client: ApiClient): Promise<number> {
	return (await client.get<UnreadCount>('/api/v1/notifications/unread-count')).count;
}

export function markRead(client: ApiClient, notificationId: string): Promise<null> {
	return client.send<null>(`/api/v1/notifications/${encodeURIComponent(notificationId)}/read`);
}

export function markAllRead(client: ApiClient): Promise<null> {
	return client.send<null>('/api/v1/notifications/read-all');
}
