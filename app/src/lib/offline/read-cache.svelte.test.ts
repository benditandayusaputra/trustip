import { beforeEach, describe, expect, it } from 'vitest';
import { DeviceReadCache, isCacheable } from './read-cache';

const envelope = (data: unknown) => ({
	code: 'OK' as const,
	message: '',
	data,
	meta: { request_id: 'r' }
});

describe('DeviceReadCache (FR-PWA-04, BR-PWA-02)', () => {
	let cache: DeviceReadCache;

	beforeEach(async () => {
		cache = new DeviceReadCache(() => new Date('2026-10-09T03:00:00Z'));
		await cache.forget();
	});

	it('gives back the last answer of a read and marks the screen as showing saved data', async () => {
		await cache.remember('/api/v1/drops?status=RECORDED', envelope([{ id: 'd1' }]));

		const stored = await cache.recall('/api/v1/drops?status=RECORDED');

		expect(stored?.data).toEqual([{ id: 'd1' }]);
		expect(cache.freshness.staleSince).toBe('2026-10-09T03:00:00.000Z');
	});

	it('clears the saved data and the stale mark when the user signs out', async () => {
		await cache.remember('/api/v1/me', envelope({ id: 'u1' }));
		await cache.recall('/api/v1/me');

		await cache.forget();

		await expect(cache.recall('/api/v1/me')).resolves.toBeUndefined();
		expect(cache.freshness.staleSince).toBeNull();
	});

	it('marks the screen fresh again after a read reaches the server', async () => {
		await cache.remember('/api/v1/stock', envelope([]));
		await cache.recall('/api/v1/stock');

		await cache.remember('/api/v1/stock', envelope([]));

		expect(cache.freshness.staleSince).toBeNull();
	});

	it('never keeps payment details or private files on the device (R-OFF-14, R-SEC-72)', async () => {
		await cache.remember('/api/v1/organizations/o1/payment-info', envelope({ account: '123' }));

		await expect(cache.recall('/api/v1/organizations/o1/payment-info')).resolves.toBeUndefined();
		expect(isCacheable('/api/v1/files/f1')).toBe(false);
		expect(isCacheable('/api/v1/drops/d1')).toBe(true);
	});
});
