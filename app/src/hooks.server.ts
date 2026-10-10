import type { Handle } from '@sveltejs/kit/hooks';
import { applySecurityHeaders } from '#lib/server/security-headers.js';
import { signInRedirectFor } from '#lib/server/app-guard.js';

const seeOther = 303;

function preload({ type, path }: { type: string; path: string }): boolean {
	return type === 'js' || type === 'css' || (type === 'font' && path.includes('/inter-latin-'));
}

export const handle: Handle = async ({ event, resolve }) => {
	const signIn = signInRedirectFor(event.url, (name) => event.cookies.get(name) !== undefined);
	const response = signIn
		? new Response(null, { status: seeOther, headers: { location: signIn } })
		: await resolve(event, { preload });
	applySecurityHeaders(response.headers);
	return response;
};
