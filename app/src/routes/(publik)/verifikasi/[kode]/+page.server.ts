import { serverApiClient } from '#lib/api/client.js';
import { verifySeal } from '#lib/api/public.js';
import { publicPageCache } from '#lib/server/public-cache.js';
import { findPublic } from '#lib/server/public-page.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, setHeaders }) => {
	const verification = await findPublic(() => verifySeal(serverApiClient(fetch), params.kode));
	setHeaders(publicPageCache);
	return { code: params.kode, verification };
};
