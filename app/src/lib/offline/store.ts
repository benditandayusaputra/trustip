const databaseName = 'trustip';
const databaseVersion = 1;
const readsStore = 'reads';

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
		request.result.createObjectStore(readsStore);
	});
	return settle(request);
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
