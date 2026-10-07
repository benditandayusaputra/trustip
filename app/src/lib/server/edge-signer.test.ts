import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { canonicalString, edgeHeaders, freshNonce, signature } from './edge-signer';

interface Vector {
	name: string;
	method: string;
	uri: string;
	timestamp: number;
	nonce: string;
	client_ip: string;
	body: string;
	canonical: string;
	signature: string;
}

interface Vectors {
	key_id: string;
	key_hex: string;
	cases: Vector[];
}

const vectors = JSON.parse(
	readFileSync(
		new URL('../../../../api/testdata/vectors/edge_signature.json', import.meta.url),
		'utf8'
	)
) as Vectors;
const key = { id: vectors.key_id, secret: Buffer.from(vectors.key_hex, 'hex') };

const requestOf = (vector: Vector) => ({
	method: vector.method,
	uri: vector.uri,
	body: new TextEncoder().encode(vector.body),
	clientIp: vector.client_ip,
	timestamp: vector.timestamp,
	nonce: vector.nonce
});

describe('edge signer against the shared vectors', () => {
	it.each(vectors.cases)('builds the canonical string for $name', (vector) => {
		expect(canonicalString(requestOf(vector))).toBe(vector.canonical);
	});

	it.each(vectors.cases)('produces the API signature for $name', (vector) => {
		expect(signature(requestOf(vector), key)).toBe(vector.signature);
	});
});

describe('edgeHeaders', () => {
	it('sends every header the API verifies', () => {
		const [vector] = vectors.cases;
		if (!vector) {
			throw new Error('missing vector');
		}

		expect(Object.keys(edgeHeaders(requestOf(vector), key)).sort()).toEqual([
			'x-edge-client-ip',
			'x-edge-key-id',
			'x-edge-nonce',
			'x-edge-signature',
			'x-edge-timestamp'
		]);
	});
});

describe('freshNonce', () => {
	it('is 16 random bytes in hex and never repeats', () => {
		const first = freshNonce();

		expect(first).toMatch(/^[0-9a-f]{32}$/u);
		expect(freshNonce()).not.toBe(first);
	});
});
