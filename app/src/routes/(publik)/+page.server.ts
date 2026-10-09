import { publicPageCache } from '#lib/server/public-cache.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ setHeaders }) => {
	setHeaders(publicPageCache);
};
