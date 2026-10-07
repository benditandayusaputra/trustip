import { signInHref } from '#lib/features/auth/links.js';

const appArea = '/app';
const sessionCookieNames = ['trustip_at', 'trustip_rt'].flatMap((name) => [name, `__Host-${name}`]);

export function signInRedirectFor(
	url: URL,
	hasCookie: (name: string) => boolean
): string | undefined {
	const isAppArea = url.pathname === appArea || url.pathname.startsWith(`${appArea}/`);
	if (!isAppArea || sessionCookieNames.some(hasCookie)) {
		return undefined;
	}
	return signInHref(url.pathname + url.search);
}
