import { browser } from '$app/env';
import { ApiError, NetworkError } from './errors';
import { readEnvelope, toPage, type Page, type SuccessEnvelope } from './envelope';
import { common as id } from '#lib/i18n/common.js';
import { deviceReadCache } from '#lib/offline/read-cache.js';

const refreshPath = '/api/v1/sessions/refresh';
const refreshLockName = 'trustip-session-refresh';

type Method = 'GET' | 'POST' | 'PATCH' | 'DELETE';

interface Call {
	method: Method;
	path: string;
	body?: unknown;
	idempotencyKey?: string;
}

export interface ReadCache {
	remember: (path: string, envelope: SuccessEnvelope<unknown>) => Promise<void>;
	recall: (path: string) => Promise<SuccessEnvelope<unknown> | undefined>;
}

export interface ClientDeps {
	fetch: typeof fetch;
	newKey: () => string;
	withRefreshLock: <T>(task: () => Promise<T>) => Promise<T>;
	readCache: ReadCache;
}

export class ApiClient {
	readonly #deps: ClientDeps;
	#refreshing: Promise<boolean> | undefined;

	constructor(deps: ClientDeps) {
		this.#deps = deps;
	}

	async get<T>(path: string): Promise<T> {
		return (await this.#call<T>({ method: 'GET', path })).data;
	}

	async list<T>(path: string): Promise<Page<T>> {
		return toPage(await this.#call<T[]>({ method: 'GET', path }));
	}

	async create<T>(path: string, body: unknown, idempotencyKey = this.#deps.newKey()): Promise<T> {
		return (await this.#call<T>({ method: 'POST', path, body, idempotencyKey })).data;
	}

	async send<T>(path: string, body?: unknown): Promise<T> {
		return (await this.#call<T>({ method: 'POST', path, body })).data;
	}

	async update<T>(path: string, body: unknown): Promise<T> {
		return (await this.#call<T>({ method: 'PATCH', path, body })).data;
	}

	async upload<T>(path: string, form: FormData): Promise<T> {
		return (await this.#call<T>({ method: 'POST', path, body: form })).data;
	}

	async #call<T>(call: Call): Promise<SuccessEnvelope<T>> {
		if (call.method !== 'GET') {
			return this.#exchange<T>(call);
		}
		try {
			const envelope = await this.#exchange<T>(call);
			void this.#deps.readCache.remember(call.path, envelope);
			return envelope;
		} catch (error) {
			if (!(error instanceof NetworkError)) {
				throw error;
			}
			return this.#recall<T>(call.path, error);
		}
	}

	async #recall<T>(path: string, failure: NetworkError): Promise<SuccessEnvelope<T>> {
		const stored = await this.#deps.readCache.recall(path);
		if (!stored) {
			throw failure;
		}
		return stored as SuccessEnvelope<T>;
	}

	async #exchange<T>(call: Call): Promise<SuccessEnvelope<T>> {
		try {
			return await readEnvelope<T>(await this.#request(call));
		} catch (error) {
			if (!this.#shouldRefresh(error, call) || !(await this.#refreshOnce())) {
				throw error;
			}
			return readEnvelope<T>(await this.#request(call));
		}
	}

	#shouldRefresh(error: unknown, call: Call): boolean {
		return error instanceof ApiError && error.isSessionProblem() && call.path !== refreshPath;
	}

	#refreshOnce(): Promise<boolean> {
		this.#refreshing ??= this.#deps
			.withRefreshLock(() => this.#request({ method: 'POST', path: refreshPath }))
			.then((response) => response.ok)
			.catch(() => false)
			.finally(() => {
				this.#refreshing = undefined;
			});
		return this.#refreshing;
	}

	async #request(call: Call): Promise<Response> {
		try {
			return await this.#deps.fetch(call.path, {
				method: call.method,
				headers: requestHeaders(call),
				body: bodyOf(call),
				credentials: 'same-origin'
			});
		} catch (cause) {
			throw new NetworkError(id.apiErrors.network, { cause });
		}
	}
}

function bodyOf(call: Call): BodyInit | null {
	if (call.body instanceof FormData) {
		return call.body;
	}
	return call.body === undefined ? null : JSON.stringify(call.body);
}

function requestHeaders(call: Call): Headers {
	const headers = new Headers({ accept: 'application/json', 'x-trustip-client': 'web' });
	if (call.body !== undefined && !(call.body instanceof FormData)) {
		headers.set('content-type', 'application/json');
	}
	if (call.idempotencyKey !== undefined) {
		headers.set('idempotency-key', call.idempotencyKey);
	}
	return headers;
}

function withWebLock<T>(task: () => Promise<T>): Promise<T> {
	return navigator.locks.request(refreshLockName, task);
}

const noReadCache: ReadCache = {
	remember: () => Promise.resolve(),
	recall: () => Promise.resolve(undefined)
};

export function serverApiClient(eventFetch: typeof fetch): ApiClient {
	return new ApiClient({
		fetch: eventFetch,
		newKey: () => crypto.randomUUID(),
		withRefreshLock: (task) => task(),
		readCache: noReadCache
	});
}

let browserClient: ApiClient | undefined;

export function apiClient(): ApiClient {
	if (!browser) {
		throw new Error('apiClient runs in the browser only, server code uses event.fetch');
	}
	browserClient ??= new ApiClient({
		fetch: (input, init) => fetch(input, init),
		newKey: () => crypto.randomUUID(),
		withRefreshLock: withWebLock,
		readCache: deviceReadCache()
	});
	return browserClient;
}
