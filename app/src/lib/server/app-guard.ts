import { signInHref } from '#lib/features/auth/links.js';

const signedInAreas = ['/app', '/onboarding'];
const sessionCookieNames = ['trustip_at', 'trustip_rt'].flatMap((name) => [name, `__Host-${name}`]);

export function signInRedirectFor(
	url: URL,
	hasCookie: (name: string) => boolean
): string | undefined {
	const isGuarded = signedInAreas.some(
		(area) => url.pathname === area || url.pathname.startsWith(`${area}/`)
	);
	if (!isGuarded || sessionCookieNames.some(hasCookie)) {
		return undefined;
	}
	return signInHref(url.pathname + url.search);
}
