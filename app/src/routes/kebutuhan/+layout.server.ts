import { publicPageCache } from '#lib/server/public-cache.js';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = ({ setHeaders }) => {
	setHeaders(publicPageCache);
};
