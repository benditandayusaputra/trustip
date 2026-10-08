import { describe, expect, it } from 'vitest';
import { ApiClient } from '#lib/api/client.js';
import type { SuccessEnvelope } from '#lib/api/envelope.js';
import type { EntryDraft, OutboxEntry } from './outbox';
import { OutboxSync, type OutboxStorage } from './sync';

interface Call {
	path: string;
	key: string | null;
	body: Record<string, unknown>;
}

type Reply = () => Response;

const ok = (data: unknown, status = 200) =>
	Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } }, { status });
const failure = (status: number, code: string) =>
	Response.json({ code, message: `pesan ${code}`, meta: { request_id: 'r9' } }, { status });
const lostConnection: Reply = () => {
	throw new TypeError('Failed to fetch');
};

const sari = {
	id: 'user-sari',
	membership: { organization_id: 'store-1', organization_type: 'STORE' }
};

function memoryStore(): OutboxStorage & { entries: Map<string, OutboxEntry> } {
	const entries = new Map<string, OutboxEntry>();
	return {
		entries,
		put: (entry) => {
			entries.set(entry.id, structuredClone(entry));
			return Promise.resolve();
		},
		get: (id) => Promise.resolve(entries.get(id)),
		listFor: (userId) =>
			Promise.resolve([...entries.values()].filter((entry) => entry.userId === userId)),
		remove: (id) => {
			entries.delete(id);
			return Promise.resolve();
		}
	};
}

function callOf(path: string, init: RequestInit): Call {
	const body = typeof init.body === 'string' ? init.body : '{}';
	return {
		path,
		key: new Headers(init.headers).get('idempotency-key'),
		body: JSON.parse(body) as Record<string, unknown>
	};
}

function answer(reply: Reply | undefined): Promise<Response> {
	try {
		return Promise.resolve(reply ? reply() : failure(404, 'NOT_FOUND'));
	} catch (error) {
		return Promise.reject(error instanceof Error ? error : new Error('reply failed'));
	}
}

function server(replies: Record<string, Reply[]>, calls: Call[]) {
	const saved = new Map<string, SuccessEnvelope<unknown>>();
	return new ApiClient({
		fetch: (input, init = {}) => {
			const path = input instanceof Request ? input.url : input.toString();
			const queue = replies[`${init.method ?? 'GET'} ${path}`] ?? [];
			if (init.method === 'POST') {
				calls.push(callOf(path, init));
			}
			return answer(queue.length > 1 ? queue.shift() : queue[0]);
		},
		newKey: () => 'unused',
		withRefreshLock: (task) => task(),
		readCache: {
			remember: (path, envelope) => {
				saved.set(path, envelope);
				return Promise.resolve();
			},
			recall: (path) => Promise.resolve(saved.get(path))
		}
	});
}

function engine(client: ApiClient, store = memoryStore()) {
	const isOnline = () => true;
	let tick = 0;
	let key = 0;
	const sync = new OutboxSync({
		client: () => client,
		store,
		withLock: (task) => task(),
		now: () => new Date(Date.parse('2026-11-10T02:00:00Z') + 1000 * tick++),
		isOnline,
		newKey: () => `key-${String(++key)}`,
		announce: () => undefined
	});
	return { sync, store };
}

const dropDraft = (partnershipId: string): EntryDraft => ({
	kind: 'DROP',
	partnershipId,
	partnerName: 'Dapur Bu Rina',
	request: { partnership_id: partnershipId, agreement_version: 1, occurred_at: null, items: [] },
	lines: []
});

const countDraft = (partnershipId: string): EntryDraft => ({
	kind: 'COUNT',
	partnershipId,
	partnerName: 'Dapur Bu Rina',
	request: {
		partnership_id: partnershipId,
		agreement_version: 1,
		counted_at: null,
		occasion: 'STORE_CLOSING',
		note: '',
		lines: []
	}
});

const me = { 'GET /api/v1/me': [() => ok(sari)] };

describe('OutboxSync (FR-PWA-06)', () => {
	it('BR-PWA-04 sends the entries of a partnership in the order they were recorded', async () => {
		const calls: Call[] = [];
		const { sync, store } = engine(
			server(
				{
					...me,
					'POST /api/v1/drops': [() => ok({ id: 'd1', number: 'TM-202611-0001' }, 201)],
					'POST /api/v1/counts': [() => ok({ id: 'c1', number: 'HS-202611-0001' }, 201)]
				},
				calls
			)
		);
		await sync.enqueue(dropDraft('p1'));
		await sync.enqueue(countDraft('p1'));

		await expect(sync.flush()).resolves.toEqual({ kind: 'done', sent: 2 });

		expect(calls.map((call) => call.path)).toEqual(['/api/v1/drops', '/api/v1/counts']);
		expect([...store.entries.values()].map((entry) => entry.document?.number)).toEqual([
			'TM-202611-0001',
			'HS-202611-0001'
		]);
	});

	it('BR-PWA-03 keeps the idempotency key when a lost connection forces a second send', async () => {
		const calls: Call[] = [];
		const { sync } = engine(
			server(
				{
					...me,
					'POST /api/v1/drops': [lostConnection, () => ok({ id: 'd1', number: 'TM-1' }, 201)]
				},
				calls
			)
		);
		const id = await sync.enqueue(dropDraft('p1'));

		await expect(sync.flush()).resolves.toEqual({ kind: 'retryLater', attempts: 1 });
		await sync.flush();

		expect(calls.map((call) => call.key)).toEqual([id, id]);
		expect(calls.map((call) => call.body.recorded_offline)).toEqual([false, true]);
		expect(calls[1]?.body.client_recorded_at).toBe('2026-11-10T02:00:00.000Z');
	});

	it('BR-PWA-06 holds the next entry of the same partner after a refusal, and sends other partners', async () => {
		const calls: Call[] = [];
		const { sync, store } = engine(
			server(
				{
					...me,
					'POST /api/v1/counts': [() => failure(409, 'STOCK_CHANGED')],
					'POST /api/v1/drops': [() => ok({ id: 'd2', number: 'TM-2' }, 201)]
				},
				calls
			)
		);
		const refused = await sync.enqueue(countDraft('p1'));
		const held = await sync.enqueue(dropDraft('p1'));
		const other = await sync.enqueue(dropDraft('p2'));

		await sync.flush();

		expect(store.entries.get(refused)).toMatchObject({
			status: 'NEEDS_REVIEW',
			problem: { code: 'STOCK_CHANGED' }
		});
		expect(store.entries.get(held)?.status).toBe('HELD');
		expect(store.entries.get(other)?.status).toBe('SENT');
	});

	it('BR-PWA-05 keeps the entries waiting when the session has ended', async () => {
		const { sync, store } = engine(
			server({ 'GET /api/v1/me': [() => failure(401, 'SESSION_EXPIRED')] }, [])
		);
		store.entries.set('e1', { ...(await queued(dropDraft('p1'))), id: 'e1' });

		await expect(sync.flush()).resolves.toEqual({ kind: 'signedOut' });

		expect(store.entries.get('e1')?.status).toBe('PENDING');
	});

	it('R-OFF-08 never sends the entries of another user', async () => {
		const calls: Call[] = [];
		const { sync, store } = engine(server(me, calls));
		store.entries.set('e1', { ...(await queued(dropDraft('p1'))), id: 'e1', userId: 'user-lain' });

		await sync.flush();

		expect(calls).toEqual([]);
	});

	it('R-OFF-10 gives a corrected entry a new idempotency key', async () => {
		const { sync, store } = engine(server(me, []));
		const id = await sync.enqueue(countDraft('p1'));
		const corrected = countDraft('p1');
		if (corrected.kind === 'COUNT') {
			corrected.request.note = 'Dihitung ulang';
		}

		await sync.revise(id, corrected);

		expect(store.entries.get(id)).toMatchObject({
			revision: 1,
			idempotencyKey: 'key-2',
			status: 'PENDING',
			request: { note: 'Dihitung ulang' }
		});
	});

	it('BR-CNT-06 R-OFF-06 sends a corrected entry at the time of the correction without moving it in the queue', async () => {
		const calls: Call[] = [];
		const { sync, store } = engine(
			server(
				{
					...me,
					'POST /api/v1/counts': [() => ok({ id: 'c1', number: 'HS-202611-0001' }, 201)]
				},
				calls
			)
		);
		const id = await sync.enqueue(countDraft('p1'));
		const recordedAt = store.entries.get(id)?.recordedAt;

		await sync.revise(id, countDraft('p1'));
		await sync.flush();

		expect(store.entries.get(id)?.recordedAt).toBe(recordedAt);
		expect(calls[0]?.body.client_recorded_at).toBe('2026-11-10T02:00:01.000Z');
	});

	it('hands a refusal on the first send back to the form and forgets the entry', async () => {
		const { sync, store } = engine(
			server({ ...me, 'POST /api/v1/drops': [() => failure(422, 'VALIDATION_FAILED')] }, [])
		);
		const id = await sync.enqueue(dropDraft('p1'));
		await sync.flush();

		const result = await sync.resultOf(id);

		expect(result).toMatchObject({ kind: 'rejected', error: { code: 'VALIDATION_FAILED' } });
		expect(store.entries.size).toBe(0);
	});
});

async function queued(draft: EntryDraft): Promise<OutboxEntry> {
	const { sync, store } = engine(server(me, []));
	const id = await sync.enqueue(draft);
	const entry = store.entries.get(id);
	if (!entry) {
		throw new Error('entry missing');
	}
	return entry;
}

describe('OutboxSync housekeeping', () => {
	it('forgets a sent entry once it is more than a day old', async () => {
		const { sync, store } = engine(server(me, []));
		const entry = await queued(dropDraft('p1'));
		store.entries.set('old', {
			...entry,
			id: 'old',
			status: 'SENT',
			recordedAt: '2026-11-08T00:00:00.000Z',
			document: { id: 'd0', number: 'TM-0' }
		});
		store.entries.set('fresh', {
			...entry,
			id: 'fresh',
			status: 'SENT',
			document: { id: 'd1', number: 'TM-1' }
		});

		await sync.flush();

		expect([...store.entries.keys()]).toEqual(['fresh']);
	});
});
