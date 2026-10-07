import { edgeHeaders, freshNonce, unixSeconds, type EdgeKey } from './edge-signer';
import { id } from '#lib/i18n/id.js';

const forwardedRequestHeaders = [
	'cookie',
	'content-type',
	'accept',
	'idempotency-key',
	'origin',
	'x-request-id',
	'x-trustip-client'
];
const forwardedResponseHeaders = [
	'content-type',
	'cache-control',
	'retry-after',
	'ratelimit-limit',
	'ratelimit-remaining',
	'ratelimit-reset',
	'x-request-id',
	'location',
	'idempotent-replay'
];
const requestIdPattern = /^[A-Za-z0-9-]{8,64}$/u;
const upstreamTimeoutMs = 15000;
const unavailableRetrySeconds = '5';
const serviceUnavailableStatus = 503;

export interface ProxyConfig {
	apiBase: string;
	key: EdgeKey;
}

export interface ProxyDeps {
	fetch: typeof fetch;
	now: () => Date;
	nonce: () => string;
	requestId: () => string;
}

export interface IncomingRequest {
	method: string;
	url: URL;
	headers: Headers;
	body: Uint8Array<ArrayBuffer>;
	clientIp: string;
}

export const defaultProxyDeps: ProxyDeps = {
	fetch: (input, init) => fetch(input, init),
	now: () => new Date(),
	nonce: freshNonce,
	requestId: () => crypto.randomUUID()
};

export async function forwardToApi(
	incoming: IncomingRequest,
	config: ProxyConfig,
	deps: ProxyDeps
): Promise<Response> {
	const requestId = requestIdOf(incoming.headers, deps);
	const uri = incoming.url.pathname + incoming.url.search;
	const signed = {
		method: incoming.method,
		uri,
		body: incoming.body,
		clientIp: incoming.clientIp,
		timestamp: unixSeconds(deps.now()),
		nonce: deps.nonce()
	};
	const headers = forwardedHeaders(incoming.headers, requestId, edgeHeaders(signed, config.key));
	try {
		const upstream = await deps.fetch(config.apiBase + uri, {
			method: incoming.method,
			headers,
			body: hasBody(incoming.method) ? incoming.body : null,
			redirect: 'manual',
			signal: AbortSignal.timeout(upstreamTimeoutMs)
		});
		return relayResponse(upstream);
	} catch {
		return unavailableResponse(requestId);
	}
}

function requestIdOf(headers: Headers, deps: ProxyDeps): string {
	const incoming = headers.get('x-request-id') ?? '';
	return requestIdPattern.test(incoming) ? incoming : deps.requestId();
}

function forwardedHeaders(
	incoming: Headers,
	requestId: string,
	signatureHeaders: Record<string, string>
): Headers {
	const headers = new Headers();
	forwardedRequestHeaders.forEach((name) => {
		const value = incoming.get(name);
		if (value !== null) {
			headers.set(name, value);
		}
	});
	headers.set('x-request-id', requestId);
	Object.entries(signatureHeaders).forEach(([name, value]) => {
		headers.set(name, value);
	});
	return headers;
}

function hasBody(method: string): boolean {
	return method !== 'GET' && method !== 'HEAD';
}

function relayResponse(upstream: Response): Response {
	const headers = new Headers();
	forwardedResponseHeaders.forEach((name) => {
		const value = upstream.headers.get(name);
		if (value !== null) {
			headers.set(name, value);
		}
	});
	upstream.headers.getSetCookie().forEach((cookie) => {
		headers.append('set-cookie', cookie);
	});
	return new Response(upstream.body, { status: upstream.status, headers });
}

function unavailableResponse(requestId: string): Response {
	const body = {
		code: 'SERVICE_UNAVAILABLE',
		message: id.apiErrors.serviceUnavailable,
		meta: { request_id: requestId }
	};
	return Response.json(body, {
		status: serviceUnavailableStatus,
		headers: { 'retry-after': unavailableRetrySeconds, 'cache-control': 'no-store' }
	});
}
