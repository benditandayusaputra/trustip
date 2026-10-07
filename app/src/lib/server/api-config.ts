import { API_INTERNAL_URL, EDGE_HMAC_ACTIVE, EDGE_HMAC_KEY_V1 } from '$app/env/private';
import type { ProxyConfig } from './proxy';

const edgeKeys: Record<string, Uint8Array | undefined> = { v1: EDGE_HMAC_KEY_V1 };

export function apiProxyConfig(): ProxyConfig {
	const secret = edgeKeys[EDGE_HMAC_ACTIVE];
	if (!secret) {
		throw new Error(`EDGE_HMAC_ACTIVE names ${EDGE_HMAC_ACTIVE} but its key is not set`);
	}
	return { apiBase: API_INTERNAL_URL, key: { id: EDGE_HMAC_ACTIVE, secret } };
}
