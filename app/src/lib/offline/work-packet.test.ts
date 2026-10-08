import { describe, expect, it } from 'vitest';
import { ApiClient, type ReadCache } from '#lib/api/client.js';
import type { SuccessEnvelope } from '#lib/api/envelope.js';
import { refreshWorkPackets } from './work-packet';

type Routes = Record<string, unknown>;

const envelope = (data: unknown, meta: object = { request_id: 'r' }) =>
	Response.json({ code: 'OK', message: '', data, meta });

function clientServing(routes: Routes, saved: Map<string, SuccessEnvelope<unknown>>) {
	const readCache: ReadCache = {
		remember: (path, answer) => {
			saved.set(path, answer);
			return Promise.resolve();
		},
		recall: (path) => Promise.resolve(saved.get(path))
	};
	return new ApiClient({
		fetch: (input) => {
			const path = input instanceof Request ? input.url : String(input);
			return Promise.resolve(
				path in routes ? envelope(routes[path]) : Response.json({}, { status: 404 })
			);
		},
		newKey: () => 'key',
		withRefreshLock: (task) => task(),
		readCache
	});
}

const memberOf = (type: 'STORE' | 'CONSIGNOR') => ({
	id: 'u1',
	membership: { organization_id: 'o1', organization_type: type, role: 'OWNER' }
});

describe('refreshWorkPackets (FR-PWA-08, R-OFF-01)', () => {
	it('reports a refresh that could not finish instead of throwing in the background', async () => {
		const client = clientServing({}, new Map());

		await expect(refreshWorkPackets(client)).resolves.toMatchObject({ kind: 'failed' });
	});

	it('keeps the agreement, catalog, and recorded stock of every active partnership on the device', async () => {
		const saved = new Map<string, SuccessEnvelope<unknown>>();
		const client = clientServing(
			{
				'/api/v1/me': memberOf('STORE'),
				'/api/v1/partnerships?status=ACTIVE': [{ id: 'p1' }, { id: 'p2' }],
				'/api/v1/partnerships/p1/agreement': { version: 3 },
				'/api/v1/partnerships/p1/catalog': [],
				'/api/v1/partnerships/p1/count-sheet': { lines: [] },
				'/api/v1/partnerships/p2/agreement': { version: 1 },
				'/api/v1/partnerships/p2/catalog': [],
				'/api/v1/partnerships/p2/count-sheet': { lines: [] }
			},
			saved
		);

		await expect(refreshWorkPackets(client)).resolves.toEqual({
			kind: 'refreshed',
			partnerships: 2
		});

		expect([...saved.keys()]).toEqual(
			expect.arrayContaining([
				'/api/v1/partnerships/p1/agreement',
				'/api/v1/partnerships/p1/catalog',
				'/api/v1/partnerships/p1/count-sheet',
				'/api/v1/partnerships/p2/count-sheet'
			])
		);
	});

	it('skips consignors, who never record drops or counts', async () => {
		const saved = new Map<string, SuccessEnvelope<unknown>>();
		const client = clientServing({ '/api/v1/me': memberOf('CONSIGNOR') }, saved);

		await expect(refreshWorkPackets(client)).resolves.toEqual({ kind: 'notAStore' });

		expect([...saved.keys()]).toEqual(['/api/v1/me']);
	});
});
