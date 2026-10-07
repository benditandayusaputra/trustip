import type { ApiClient } from './client';
import type { Page } from './envelope';
import type { components } from './schema';

export type Product = components['schemas']['Product'];
export type ProductRequest = components['schemas']['ProductRequest'];

export interface ProductSearch {
	text: string;
	categoryId: string;
	status: Product['status'];
}

const productsPath = '/api/v1/products';

function productPath(productId: string): string {
	return `${productsPath}/${encodeURIComponent(productId)}`;
}

export function listProducts(
	client: ApiClient,
	search: ProductSearch,
	cursor: string | null
): Promise<Page<Product>> {
	const query = new URLSearchParams({ status: search.status });
	if (search.text !== '') {
		query.set('q', search.text);
	}
	if (search.categoryId !== '') {
		query.set('category_id', search.categoryId);
	}
	if (cursor) {
		query.set('cursor', cursor);
	}
	return client.list<Product>(`${productsPath}?${query.toString()}`);
}

export function getProduct(client: ApiClient, productId: string): Promise<Product> {
	return client.get<Product>(productPath(productId));
}

export function createProduct(client: ApiClient, request: ProductRequest): Promise<Product> {
	return client.send<Product>(productsPath, request);
}

export function updateProduct(
	client: ApiClient,
	productId: string,
	request: ProductRequest
): Promise<Product> {
	return client.update<Product>(productPath(productId), request);
}

export function archiveProduct(client: ApiClient, productId: string): Promise<Product> {
	return client.send<Product>(`${productPath(productId)}/archive`);
}

export function reactivateProduct(client: ApiClient, productId: string): Promise<Product> {
	return client.send<Product>(`${productPath(productId)}/reactivate`);
}

export function setProductSeeking(
	client: ApiClient,
	productId: string,
	isSeekingStores: boolean
): Promise<Product> {
	return client.update<Product>(`${productPath(productId)}/seeking`, {
		is_seeking_stores: isSeekingStores
	});
}
