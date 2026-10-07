import type { Handle } from '@sveltejs/kit/hooks';
import { applySecurityHeaders } from '#lib/server/security-headers.js';

export const handle: Handle = async ({ event, resolve }) => {
	const response = await resolve(event);
	applySecurityHeaders(response.headers);
	return response;
};
