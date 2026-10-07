import { createHash, createHmac, randomBytes } from 'node:crypto';

const canonicalVersion = 'v1';
const nonceBytes = 16;
const millisecondsPerSecond = 1000;

export interface EdgeKey {
	id: string;
	secret: Uint8Array;
}

export interface SignableRequest {
	method: string;
	uri: string;
	body: Uint8Array;
	clientIp: string;
	timestamp: number;
	nonce: string;
}

export function canonicalString(request: SignableRequest): string {
	const bodyHash = createHash('sha256').update(request.body).digest('hex');
	return [
		canonicalVersion,
		request.method,
		request.uri,
		String(request.timestamp),
		request.nonce,
		request.clientIp,
		bodyHash
	].join('\n');
}

export function signature(request: SignableRequest, key: EdgeKey): string {
	return createHmac('sha256', key.secret).update(canonicalString(request)).digest('hex');
}

export function edgeHeaders(request: SignableRequest, key: EdgeKey): Record<string, string> {
	return {
		'x-edge-key-id': key.id,
		'x-edge-timestamp': String(request.timestamp),
		'x-edge-nonce': request.nonce,
		'x-edge-client-ip': request.clientIp,
		'x-edge-signature': signature(request, key)
	};
}

export function freshNonce(): string {
	return randomBytes(nonceBytes).toString('hex');
}

export function unixSeconds(date: Date): number {
	return Math.floor(date.getTime() / millisecondsPerSecond);
}
