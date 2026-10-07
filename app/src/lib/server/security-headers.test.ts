import { describe, expect, it } from 'vitest';
import { applySecurityHeaders } from './security-headers';

describe('applySecurityHeaders', () => {
	it('sets every header from the security table', () => {
		const headers = new Headers();

		applySecurityHeaders(headers);

		expect(Object.fromEntries(headers)).toEqual({
			'strict-transport-security': 'max-age=63072000; includeSubDomains',
			'x-content-type-options': 'nosniff',
			'referrer-policy': 'strict-origin-when-cross-origin',
			'permissions-policy': 'geolocation=(self), camera=(), microphone=(), payment=()',
			'cross-origin-opener-policy': 'same-origin',
			'x-frame-options': 'DENY',
			'cross-origin-resource-policy': 'same-origin'
		});
	});

	it('replaces a weaker value set earlier', () => {
		const headers = new Headers({ 'X-Frame-Options': 'SAMEORIGIN' });

		applySecurityHeaders(headers);

		expect(headers.get('X-Frame-Options')).toBe('DENY');
	});
});
