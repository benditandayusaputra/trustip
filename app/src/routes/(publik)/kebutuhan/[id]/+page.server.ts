import { serverApiClient } from '#lib/api/client.js';
import { getPublicNeed } from '#lib/api/public.js';
import { publicPageCache } from '#lib/server/public-cache.js';
import { readPublic } from '#lib/server/public-page.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, setHeaders }) => {
	const need = await readPublic(() => getPublicNeed(serverApiClient(fetch), params.id));
	setHeaders(publicPageCache);
	return { need };
};
