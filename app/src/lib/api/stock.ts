import type { ApiClient } from './client';
import type { components } from './schema';

export type ProductStock = components['schemas']['ProductStock'];
export type StockHolding = components['schemas']['StockHolding'];
export type StockCard = components['schemas']['StockCard'];
export type BatchFlag = components['schemas']['BatchFlag'];

function partnershipStockPath(partnershipId: string): string {
	return `/api/v1/partnerships/${encodeURIComponent(partnershipId)}/stock`;
}

export async function listStock(client: ApiClient): Promise<StockHolding[]> {
	return (await client.list<StockHolding>('/api/v1/stock')).items;
}

export async function getPartnershipStock(
	client: ApiClient,
	partnershipId: string
): Promise<ProductStock[]> {
	return (await client.list<ProductStock>(partnershipStockPath(partnershipId))).items;
}

export interface CardKey {
	partnershipId: string;
	productId: string;
}

export function getStockCard(
	client: ApiClient,
	key: CardKey,
	cursor: string | null
): Promise<StockCard> {
	const suffix = cursor ? `?cursor=${encodeURIComponent(cursor)}` : '';
	const productPath = encodeURIComponent(key.productId);
	return client.get<StockCard>(
		`${partnershipStockPath(key.partnershipId)}/${productPath}/card${suffix}`
	);
}
