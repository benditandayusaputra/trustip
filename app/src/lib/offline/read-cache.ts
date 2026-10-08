import { browser } from '$app/env';
import type { ReadCache } from '#lib/api/client.js';
import type { SuccessEnvelope } from '#lib/api/envelope.js';
import { Freshness } from './freshness.svelte';
import { cacheStore } from './store';

const privatePaths = [/\/payment-info$/u, /^\/api\/v1\/files(\/|$)/u];

export function isCacheable(path: string): boolean {
	const [pathname = ''] = path.split('?');
	return !privatePaths.some((pattern) => pattern.test(pathname));
}

export class DeviceReadCache implements ReadCache {
	readonly freshness = new Freshness();
	readonly #now: () => Date;

	constructor(now: () => Date) {
		this.#now = now;
	}

	async remember(path: string, envelope: SuccessEnvelope<unknown>): Promise<void> {
		this.freshness.markFresh();
		if (isCacheable(path)) {
			await cacheStore.put(path, { body: envelope, storedAt: this.#now().toISOString() });
		}
	}

	async recall(path: string): Promise<SuccessEnvelope<unknown> | undefined> {
		const read = isCacheable(path) ? await cacheStore.get(path) : undefined;
		if (read) {
			this.freshness.markStale(read.storedAt);
		}
		return read?.body as SuccessEnvelope<unknown> | undefined;
	}

	async forget(): Promise<void> {
		await cacheStore.clear();
		this.freshness.markFresh();
	}
}

let browserCache: DeviceReadCache | undefined;

export function deviceReadCache(): DeviceReadCache {
	if (!browser) {
		throw new Error('the read cache lives in the browser only');
	}
	browserCache ??= new DeviceReadCache(() => new Date());
	return browserCache;
}
