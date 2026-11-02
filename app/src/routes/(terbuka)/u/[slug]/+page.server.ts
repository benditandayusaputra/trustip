import { serverApiClient } from '#lib/api/client.js';
import { getPublicConsignor } from '#lib/api/public.js';
import { publicPageCache } from '#lib/server/public-cache.js';
import { readPublic } from '#lib/server/public-page.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, setHeaders }) => {
	const consignor = await readPublic(() => getPublicConsignor(serverApiClient(fetch), params.slug));
	setHeaders(publicPageCache);
	return { consignor };
};
