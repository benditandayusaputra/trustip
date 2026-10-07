import type { Product } from '#lib/api/products.js';

export function publicProductUrl(
	origin: string,
	product: Pick<Product, 'consignor_slug' | 'sku'>
): string {
	const slug = encodeURIComponent(product.consignor_slug);
	return `${origin}/p/${slug}/${encodeURIComponent(product.sku)}`;
}
