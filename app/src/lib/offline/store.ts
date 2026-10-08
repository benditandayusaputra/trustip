import type { OutboxEntry } from './outbox';
import type { OutboxStorage } from './sync';

const databaseName = 'trustip';
const databaseVersion = 2;
const readsStore = 'reads';
const outboxName = 'outbox';
const userIndex = 'userId';

let opening: Promise<IDBDatabase> | undefined;

function settle<T>(request: IDBRequest<T>): Promise<T> {
	return new Promise((resolve, reject) => {
		request.addEventListener('success', () => {
			resolve(request.result);
		});
		request.addEventListener('error', () => {
			reject(request.error ?? new Error('IndexedDB request failed'));
		});
	});
}

function openDatabase(): Promise<IDBDatabase> {
	const request = indexedDB.open(databaseName, databaseVersion);
	request.addEventListener('upgradeneeded', () => {
		createStores(request.result);
	});
	return settle(request);
}

function createStores(connection: IDBDatabase): void {
	if (!connection.objectStoreNames.contains(readsStore)) {
		connection.createObjectStore(readsStore);
	}
	if (!connection.objectStoreNames.contains(outboxName)) {
		connection.createObjectStore(outboxName, { keyPath: 'id' }).createIndex(userIndex, userIndex);
	}
}

function database(): Promise<IDBDatabase> {
	opening ??= openDatabase();
	return opening;
}

async function run<T>(
	storeName: string,
	mode: IDBTransactionMode,
	work: (store: IDBObjectStore) => IDBRequest<T>
): Promise<T> {
	const connection = await database();
	return settle(work(connection.transaction(storeName, mode).objectStore(storeName)));
}

export interface StoredRead {
	body: unknown;
	storedAt: string;
}

export const cacheStore = {
	async put(key: string, read: StoredRead): Promise<void> {
		await run(readsStore, 'readwrite', (store) => store.put(read, key));
	},
	get(key: string): Promise<StoredRead | undefined> {
		return run(
			readsStore,
			'readonly',
			(store) => store.get(key) as IDBRequest<StoredRead | undefined>
		);
	},
	async clear(): Promise<void> {
		await run(readsStore, 'readwrite', (store) => store.clear());
	}
};

function plainCopy(entry: OutboxEntry): OutboxEntry {
	return JSON.parse(JSON.stringify(entry)) as OutboxEntry;
}

export const outboxStore: OutboxStorage = {
	async put(entry) {
		await run(outboxName, 'readwrite', (store) => store.put(plainCopy(entry)));
	},
	get(id) {
		return run(
			outboxName,
			'readonly',
			(store) => store.get(id) as IDBRequest<OutboxEntry | undefined>
		);
	},
	listFor(userId) {
		return run(
			outboxName,
			'readonly',
			(store) => store.index(userIndex).getAll(userId) as IDBRequest<OutboxEntry[]>
		);
	},
	async remove(id) {
		await run(outboxName, 'readwrite', (store) => store.delete(id));
	}
};
