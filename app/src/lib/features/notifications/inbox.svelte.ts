import type { ApiClient } from '#lib/api/client.js';
import {
	countUnread,
	listNotifications,
	markAllRead,
	markRead,
	type Notification
} from '#lib/api/notifications.js';

export type InboxStatus = 'loading' | 'ready' | 'failed';
export const unreadChangedEvent = 'trustip:unread-changed';

function announceUnreadChange(): void {
	window.dispatchEvent(new Event(unreadChangedEvent));
}

export class Inbox {
	items = $state<Notification[]>([]);
	status = $state<InboxStatus>('loading');
	failure = $state('');
	nextCursor = $state<string | null>(null);
	busy = $state(false);
	unread = $state(0);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(): Promise<void> {
		this.status = 'loading';
		try {
			const [page, unread] = await Promise.all([
				listNotifications(this.#client(), null),
				countUnread(this.#client()).catch(() => 0)
			]);
			this.items = page.items;
			this.unread = Math.max(unread, page.items.filter((item) => !item.is_read).length);
			this.nextCursor = page.hasMore ? page.nextCursor : null;
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	async loadMore(): Promise<void> {
		if (this.nextCursor === null) {
			return;
		}
		this.busy = true;
		try {
			const page = await listNotifications(this.#client(), this.nextCursor);
			this.items = [...this.items, ...page.items];
			this.nextCursor = page.hasMore ? page.nextCursor : null;
		} finally {
			this.busy = false;
		}
	}

	async open(notification: Notification): Promise<void> {
		if (notification.is_read) {
			return;
		}
		await markRead(this.#client(), notification.id);
		this.items = this.items.map((item) =>
			item.id === notification.id ? { ...item, is_read: true } : item
		);
		this.unread = Math.max(0, this.unread - 1);
		announceUnreadChange();
	}

	async readAll(): Promise<void> {
		this.busy = true;
		try {
			await markAllRead(this.#client());
			this.items = this.items.map((item) => ({ ...item, is_read: true }));
			this.unread = 0;
			announceUnreadChange();
		} finally {
			this.busy = false;
		}
	}
}

export class UnreadCounter {
	count = $state(0);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async refresh(): Promise<void> {
		this.count = await countUnread(this.#client()).catch(() => this.count);
	}
}
