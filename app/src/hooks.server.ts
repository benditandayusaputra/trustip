import type { Handle } from '@sveltejs/kit/hooks';
import { applySecurityHeaders } from '#lib/server/security-headers.js';
import { signInRedirectFor } from '#lib/server/app-guard.js';

const seeOther = 303;

export const handle: Handle = async ({ event, resolve }) => {
	const signIn = signInRedirectFor(event.url, (name) => event.cookies.get(name) !== undefined);
	const response = signIn
		? new Response(null, { status: seeOther, headers: { location: signIn } })
		: await resolve(event);
	applySecurityHeaders(response.headers);
	return response;
};
