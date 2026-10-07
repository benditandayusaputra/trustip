import { describe, expect, it } from 'vitest';
import { contentSecurityPolicy } from './csp.config';

describe('contentSecurityPolicy', () => {
	const policy = contentSecurityPolicy({
		files: 'https://api.trustip.id/files',
		stream: 'https://api.trustip.id/api/v1/stream'
	});

	it('never allows inline or evaluated scripts', () => {
		expect(policy['script-src']).toEqual(['self', 'https://challenges.cloudflare.com']);
	});

	it('lets only the captcha widget run in a frame', () => {
		expect(policy['frame-src']).toEqual(['https://challenges.cloudflare.com']);
	});

	it('opens only style attributes for the framework route announcer', () => {
		expect(policy['style-src']).toEqual(['self']);
		expect(policy['style-src-attr']).toEqual(['unsafe-inline']);
	});

	it('allows API file images and the event stream by origin only', () => {
		expect(policy['img-src']).toContain('https://api.trustip.id');
		expect(policy['connect-src']).toEqual(['self', 'https://api.trustip.id']);
	});

	it('forbids framing and plugins', () => {
		expect(policy['frame-ancestors']).toEqual(['none']);
		expect(policy['object-src']).toEqual(['none']);
	});

	it('leaves out origins that are not configured', () => {
		expect(contentSecurityPolicy({})['connect-src']).toEqual(['self']);
	});
});
