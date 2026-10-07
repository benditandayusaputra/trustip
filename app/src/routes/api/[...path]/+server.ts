import { apiProxyConfig } from '#lib/server/api-config.js';
import { defaultProxyDeps, forwardToApi } from '#lib/server/proxy.js';
import type { RequestHandler } from './$types';

export const fallback: RequestHandler = async ({ request, url, getClientAddress }) =>
	forwardToApi(
		{
			method: request.method,
			url,
			headers: request.headers,
			body: new Uint8Array(await request.arrayBuffer()),
			clientIp: getClientAddress()
		},
		apiProxyConfig(),
		defaultProxyDeps
	);
