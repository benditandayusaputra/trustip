import { createHmac } from 'node:crypto';
import { describe, expect, it } from 'vitest';
import { canonicalString } from './edge-signer';
import { forwardToApi, type IncomingRequest, type ProxyDeps } from './proxy';

const config = {
	apiBase: 'http://api.internal:8080',
	key: { id: 'v1', secret: new Uint8Array(32).fill(9) }
};
const signedAt = new Date('2026-10-19T01:00:00Z');

interface Captured {
	url: string;
	init: RequestInit;
}

function fakeDeps(respond: () => Response, captured: Captured[]): ProxyDeps {
	return {
		fetch: (input, init = {}) => {
			captured.push({ url: input instanceof Request ? input.url : input.toString(), init });
			return Promise.resolve(respond());
		},
		now: () => signedAt,
		nonce: () => '00112233445566778899aabbccddeeff',
		requestId: () => '0192f0c4-8a11-7d55-b0aa-3c1f2e4d5a6b'
	};
}

function incoming(overrides: Partial<IncomingRequest> = {}): IncomingRequest {
	return {
		method: 'GET',
		url: new URL('https://trustip.id/api/v1/meta'),
		headers: new Headers({
			cookie: 'trustip_at=abc',
			'x-edge-signature': 'forged',
			'x-edge-client-ip': '6.6.6.6',
			'x-forwarded-for': '6.6.6.6',
			authorization: 'Bearer stolen'
		}),
		body: new Uint8Array(),
		clientIp: '203.0.113.7',
		...overrides
	};
}

const sentHeaders = (captured: Captured[]) => new Headers(captured[0]?.init.headers);

describe('forwardToApi', () => {
	it('forwards only allowed headers and replaces edge headers with its own', async () => {
		const captured: Captured[] = [];

		await forwardToApi(
			incoming(),
			config,
			fakeDeps(() => Response.json({}), captured)
		);

		const headers = sentHeaders(captured);
		expect(headers.get('cookie')).toBe('trustip_at=abc');
		expect(headers.get('authorization')).toBeNull();
		expect(headers.get('x-forwarded-for')).toBeNull();
		expect(headers.get('x-edge-client-ip')).toBe('203.0.113.7');
		expect(headers.get('x-edge-signature')).not.toBe('forged');
	});

	it('signs the exact path, query, client address, and body the API receives', async () => {
		const captured: Captured[] = [];
		const body = new TextEncoder().encode('{"qty":2}');
		const request = incoming({
			method: 'POST',
			url: new URL('https://trustip.id/api/v1/counts?draft=1'),
			headers: new Headers({ origin: 'https://trustip.id', 'x-trustip-client': 'web' }),
			body
		});

		await forwardToApi(
			request,
			config,
			fakeDeps(() => Response.json({}), captured)
		);

		const expected = createHmac('sha256', config.key.secret)
			.update(
				canonicalString({
					method: 'POST',
					uri: '/api/v1/counts?draft=1',
					body,
					clientIp: '203.0.113.7',
					timestamp: signedAt.getTime() / 1000,
					nonce: '00112233445566778899aabbccddeeff'
				})
			)
			.digest('hex');
		expect(captured[0]?.url).toBe('http://api.internal:8080/api/v1/counts?draft=1');
		expect(sentHeaders(captured).get('x-edge-signature')).toBe(expected);
		expect(captured[0]?.init.body).toBe(body);
	});

	it('creates a request id when the browser sends none', async () => {
		const captured: Captured[] = [];

		await forwardToApi(
			incoming(),
			config,
			fakeDeps(() => Response.json({}), captured)
		);

		expect(sentHeaders(captured).get('x-request-id')).toBe('0192f0c4-8a11-7d55-b0aa-3c1f2e4d5a6b');
	});

	it('R-SEC-71 relays the inline disposition of a private file', async () => {
		const upstream = () =>
			new Response('png', {
				headers: {
					'content-type': 'image/png',
					'content-disposition': 'inline',
					'cache-control': 'no-store'
				}
			});

		const response = await forwardToApi(incoming(), config, fakeDeps(upstream, []));

		expect(response.headers.get('content-disposition')).toBe('inline');
		expect(response.headers.get('cache-control')).toBe('no-store');
	});

	it('passes every Set-Cookie through unchanged', async () => {
		const upstream = () => {
			const headers = new Headers({ 'content-type': 'application/json', 'x-internal': 'secret' });
			headers.append('set-cookie', '__Host-trustip_at=a; Path=/; HttpOnly; Secure; SameSite=Lax');
			headers.append('set-cookie', '__Host-trustip_rt=b; Path=/; HttpOnly; Secure; SameSite=Lax');
			return new Response('{}', { status: 201, headers });
		};

		const response = await forwardToApi(incoming(), config, fakeDeps(upstream, []));

		expect(response.status).toBe(201);
		expect(response.headers.getSetCookie()).toEqual([
			'__Host-trustip_at=a; Path=/; HttpOnly; Secure; SameSite=Lax',
			'__Host-trustip_rt=b; Path=/; HttpOnly; Secure; SameSite=Lax'
		]);
		expect(response.headers.get('x-internal')).toBeNull();
	});

	it('answers 503 in the standard envelope when the API cannot be reached', async () => {
		const deps = fakeDeps(() => Response.json({}), []);
		deps.fetch = () => Promise.reject(new TypeError('fetch failed'));

		const response = await forwardToApi(incoming(), config, deps);

		expect(response.status).toBe(503);
		expect(response.headers.get('retry-after')).toBe('5');
		expect(await response.json()).toEqual({
			code: 'SERVICE_UNAVAILABLE',
			message: 'Layanan sedang sibuk. Coba lagi sebentar lagi.',
			meta: { request_id: '0192f0c4-8a11-7d55-b0aa-3c1f2e4d5a6b' }
		});
	});

	it('R-SEC-21 refuses writes that do not come from this app before calling the API', async () => {
		const attempts = [
			new Headers({ origin: 'https://evil.example', 'x-trustip-client': 'web' }),
			new Headers({ 'x-trustip-client': 'web' }),
			new Headers({ origin: 'https://trustip.id' })
		];
		for (const headers of attempts) {
			const captured: Captured[] = [];

			const response = await forwardToApi(
				incoming({ method: 'POST', url: new URL('https://trustip.id/api/v1/sessions'), headers }),
				config,
				fakeDeps(() => Response.json({}), captured)
			);

			expect(response.status).toBe(403);
			expect(((await response.json()) as { code: string }).code).toBe('CROSS_SITE_REJECTED');
			expect(captured).toHaveLength(0);
		}
	});

	it('lets same-origin writes from the web client through', async () => {
		const captured: Captured[] = [];
		const headers = new Headers({ origin: 'https://trustip.id', 'x-trustip-client': 'web' });

		await forwardToApi(
			incoming({ method: 'POST', url: new URL('https://trustip.id/api/v1/sessions'), headers }),
			config,
			fakeDeps(() => Response.json({}), captured)
		);

		expect(captured).toHaveLength(1);
	});
});
