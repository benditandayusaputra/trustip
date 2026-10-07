import { describe, expect, it } from 'vitest';
import { signInRedirectFor } from './app-guard';

const noCookies = () => false;

describe('signInRedirectFor (FR-AUTH-06)', () => {
	it('sends a visitor without a session from the app area to sign in, keeping the address', () => {
		const url = new URL('https://trustip.id/app/titipan?mitra=7');

		expect(signInRedirectFor(url, noCookies)).toBe('/masuk?lanjut=%2Fapp%2Ftitipan%3Fmitra%3D7');
	});

	it('guards the app root as well', () => {
		expect(signInRedirectFor(new URL('https://trustip.id/app'), noCookies)).toBe(
			'/masuk?lanjut=%2Fapp'
		);
	});

	it.each(['trustip_at', 'trustip_rt', '__Host-trustip_at', '__Host-trustip_rt'])(
		'lets the page load when %s is present so the client can refresh',
		(name) => {
			const url = new URL('https://trustip.id/app');

			expect(signInRedirectFor(url, (cookie) => cookie === name)).toBeUndefined();
		}
	);

	it('FR-AUTH-05 guards onboarding, which needs a signed-in user', () => {
		expect(signInRedirectFor(new URL('https://trustip.id/onboarding'), noCookies)).toBe(
			'/masuk?lanjut=%2Fonboarding'
		);
	});

	it.each(['/', '/masuk', '/application', '/apps/x', '/onboardingx'])('leaves %s alone', (path) => {
		expect(signInRedirectFor(new URL(`https://trustip.id${path}`), noCookies)).toBeUndefined();
	});
});
