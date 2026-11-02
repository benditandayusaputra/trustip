import { serverApiClient } from '#lib/api/client.js';
import { getPublicProduct } from '#lib/api/public.js';
import { publicPageCache } from '#lib/server/public-cache.js';
import { readPublic } from '#lib/server/public-page.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch, setHeaders }) => {
	const product = await readPublic(() =>
		getPublicProduct(serverApiClient(fetch), { slug: params.slug, sku: params.sku })
	);
	setHeaders(publicPageCache);
	return { product, consignorSlug: params.slug };
};
