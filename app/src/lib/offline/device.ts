import { browser } from '$app/env';
import { getAccount } from '#lib/api/account.js';
import { apiClient } from '#lib/api/client.js';
import { announceOutboxChange } from './announcer';
import { isWaiting, type OutboxEntry } from './outbox';
import { OutboxScheduler } from './scheduler';
import { outboxStore } from './store';
import { OutboxSync } from './sync';
import { refreshWorkPackets } from './work-packet';

const outboxLockName = 'trustip-outbox';

let scheduler: OutboxScheduler | undefined;

function withOutboxLock<T>(task: () => Promise<T>): Promise<T> {
	return navigator.locks.request(outboxLockName, task);
}

export function outboxScheduler(): OutboxScheduler {
	if (!browser) {
		throw new Error('the outbox lives in the browser only');
	}
	scheduler ??= new OutboxScheduler({
		sync: new OutboxSync({
			client: apiClient,
			store: outboxStore,
			withLock: withOutboxLock,
			now: () => new Date(),
			isOnline: () => navigator.onLine,
			newKey: () => crypto.randomUUID(),
			announce: announceOutboxChange
		}),
		afterSent: () => void refreshWorkPackets(apiClient()),
		keepStorage: () => void navigator.storage.persist(),
		timers: {
			set: (task, delayMs) => window.setTimeout(task, delayMs),
			clear: (handle) => {
				window.clearTimeout(handle);
			}
		},
		random: Math.random
	});
	return scheduler;
}

export async function ownEntries(): Promise<OutboxEntry[]> {
	const account = await getAccount(apiClient());
	return outboxStore.listFor(account.id);
}

export async function findOwnEntry(entryId: string): Promise<OutboxEntry | undefined> {
	return (await ownEntries()).find((entry) => entry.id === entryId);
}

export async function waitingEntries(): Promise<OutboxEntry[]> {
	return (await ownEntries()).filter(isWaiting);
}

export async function discardOwnEntries(): Promise<void> {
	for (const entry of await ownEntries()) {
		await outboxStore.remove(entry.id);
	}
	announceOutboxChange();
}
