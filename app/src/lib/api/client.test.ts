import { describe, expect, it } from 'vitest';
import { ApiClient } from './client';
import { ApiError, NetworkError } from './errors';

interface Sent {
	path: string;
	init: RequestInit;
}

const ok = (data: unknown, meta: object = { request_id: 'req-1' }) =>
	Response.json({ code: 'OK', message: '', data, meta });

const failure = (status: number, code: string, extra: object = {}) =>
	Response.json(
		{ code, message: `pesan ${code}`, meta: { request_id: 'req-9' }, ...extra },
		{ status }
	);

function clientWith(responses: (() => Response)[], sent: Sent[] = []) {
	const queue = [...responses];
	return new ApiClient({
		fetch: (input, init = {}) => {
			sent.push({ path: input instanceof Request ? input.url : input.toString(), init });
			const next = queue.shift();
			return next ? Promise.resolve(next()) : Promise.reject(new TypeError('no response queued'));
		},
		newKey: () => 'generated-key',
		withRefreshLock: (task) => task()
	});
}

const headerOf = (sent: Sent | undefined, name: string) =>
	new Headers(sent?.init.headers).get(name);

describe('ApiClient reading', () => {
	it('unwraps data from the envelope', async () => {
		const client = clientWith([() => ok({ version: 'abc' })]);

		await expect(client.get('/api/v1/meta')).resolves.toEqual({ version: 'abc' });
	});

	it('turns list meta into a page', async () => {
		const client = clientWith([
			() => ok([{ id: '1' }], { request_id: 'r', next_cursor: 'c2', has_more: true })
		]);

		await expect(client.list('/api/v1/drops')).resolves.toEqual({
			items: [{ id: '1' }],
			nextCursor: 'c2',
			hasMore: true
		});
	});

	it('marks every request as coming from the web client', async () => {
		const sent: Sent[] = [];

		await clientWith([() => ok(null)], sent).get('/api/v1/meta');

		expect(headerOf(sent[0], 'x-trustip-client')).toBe('web');
	});
});

describe('ApiClient errors', () => {
	it('maps the error envelope to a typed ApiError', async () => {
		const fields = [{ field: 'items[0].qty', code: 'MIN', message: 'Jumlah minimal 1.' }];
		const client = clientWith([() => failure(422, 'VALIDATION_FAILED', { errors: fields })]);

		const error = await client.get('/api/v1/x').catch((caught: unknown) => caught);

		expect(error).toBeInstanceOf(ApiError);
		expect(error).toMatchObject({
			status: 422,
			code: 'VALIDATION_FAILED',
			requestId: 'req-9',
			fields
		});
	});

	it('reads Retry-After on rate limits', async () => {
		const limited = () =>
			Response.json(
				{ code: 'RATE_LIMITED', message: 'x', meta: { request_id: 'r' } },
				{ status: 429, headers: { 'retry-after': '30' } }
			);

		const error = await clientWith([limited])
			.get('/api/v1/x')
			.catch((caught: unknown) => caught);

		expect(error).toMatchObject({ code: 'RATE_LIMITED', retryAfterSeconds: 30 });
	});

	it('reports an unreadable answer without leaking its body', async () => {
		const gateway = () => new Response('<html>Bad Gateway</html>', { status: 502 });

		const error = await clientWith([gateway])
			.get('/api/v1/x')
			.catch((caught: unknown) => caught);

		expect(error).toMatchObject({ status: 502, code: 'INTERNAL' });
		expect((error as Error).message).not.toContain('html');
	});

	it('reports a lost connection as a NetworkError', async () => {
		const error = await clientWith([])
			.get('/api/v1/x')
			.catch((caught: unknown) => caught);

		expect(error).toBeInstanceOf(NetworkError);
	});
});

describe('ApiClient writing', () => {
	it('sends an idempotency key and a JSON body when creating', async () => {
		const sent: Sent[] = [];

		await clientWith([() => ok({ id: '7' })], sent).create('/api/v1/drops', { qty: 2 });

		expect(headerOf(sent[0], 'idempotency-key')).toBe('generated-key');
		expect(headerOf(sent[0], 'content-type')).toBe('application/json');
		expect(sent[0]?.init.body).toBe('{"qty":2}');
	});

	it('reuses the key given by the offline queue', async () => {
		const sent: Sent[] = [];

		await clientWith([() => ok({})], sent).create('/api/v1/counts', {}, 'outbox-entry-1');

		expect(headerOf(sent[0], 'idempotency-key')).toBe('outbox-entry-1');
	});

	it('does not add an idempotency key to plain actions', async () => {
		const sent: Sent[] = [];

		await clientWith([() => ok(null)], sent).send('/api/v1/drops/7/confirm');

		expect(headerOf(sent[0], 'idempotency-key')).toBeNull();
	});
});

describe('ApiClient session refresh', () => {
	it('refreshes once on an expired session and retries the request', async () => {
		const sent: Sent[] = [];
		const client = clientWith(
			[() => failure(401, 'SESSION_EXPIRED'), () => ok(null), () => ok({ name: 'Warung Bu Sri' })],
			sent
		);

		await expect(client.get('/api/v1/me')).resolves.toEqual({ name: 'Warung Bu Sri' });
		expect(sent.map((request) => request.path)).toEqual([
			'/api/v1/me',
			'/api/v1/sessions/refresh',
			'/api/v1/me'
		]);
	});

	it('shares one refresh between requests that fail together', async () => {
		const sent: Sent[] = [];
		const client = clientWith(
			[
				() => failure(401, 'SESSION_EXPIRED'),
				() => failure(401, 'SESSION_EXPIRED'),
				() => ok(null),
				() => ok('a'),
				() => ok('b')
			],
			sent
		);

		await Promise.all([client.get('/api/v1/a'), client.get('/api/v1/b')]);

		expect(sent.filter((request) => request.path.endsWith('/refresh'))).toHaveLength(1);
	});

	it('gives up with the original error when the refresh is rejected', async () => {
		const client = clientWith([
			() => failure(401, 'SESSION_EXPIRED'),
			() => failure(401, 'SESSION_EXPIRED')
		]);

		await expect(client.get('/api/v1/me')).rejects.toMatchObject({ code: 'SESSION_EXPIRED' });
	});

	it('never refreshes for a bad edge signature', async () => {
		const sent: Sent[] = [];
		const client = clientWith([() => failure(401, 'EDGE_SIGNATURE_INVALID')], sent);

		await expect(client.get('/api/v1/me')).rejects.toMatchObject({
			code: 'EDGE_SIGNATURE_INVALID'
		});
		expect(sent).toHaveLength(1);
	});
});

describe('ApiClient uploading', () => {
	it('R-SEC-22 sends multipart form data and lets the browser set the boundary', async () => {
		const sent: Sent[] = [];
		const form = new FormData();
		form.set('purpose', 'LOGO');

		await clientWith([() => ok({ id: 'file-1' })], sent).upload('/api/v1/files', form);

		expect(sent[0]?.init.body).toBe(form);
		expect(headerOf(sent[0], 'content-type')).toBeNull();
		expect(headerOf(sent[0], 'x-trustip-client')).toBe('web');
	});
});
