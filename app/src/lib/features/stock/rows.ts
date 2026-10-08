import type { BatchFlag, ProductStock, StockHolding } from '#lib/api/stock.js';
import { formatBusinessDate, formatInteger } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';

export type FlagTone = 'success' | 'pending' | 'danger';

export function flagToneOf(flag: BatchFlag): FlagTone {
	const tones: Record<BatchFlag, FlagTone> = {
		SAFE: 'success',
		NEAR_EXPIRY: 'pending',
		EXPIRED: 'danger'
	};
	return tones[flag];
}

export interface BatchRow extends Record<string, string> {
	id: string;
	batch: string;
	expiry: string;
	age: string;
	remaining: string;
	flag: string;
}

export function batchRows(product: ProductStock): BatchRow[] {
	const text = id.stock;
	return product.batches.map((batch) => ({
		id: batch.id,
		batch: batch.batch_code,
		expiry: batch.expiry_date === null ? text.noExpiry : formatBusinessDate(batch.expiry_date),
		age: text.ageDays(batch.age_days),
		remaining: formatInteger(batch.balance),
		flag: text.flags[batch.flag]
	}));
}

export interface HoldingGroup {
	productId: string;
	productName: string;
	unit: string;
	balance: number;
	holdings: StockHolding[];
}

export function groupByProduct(holdings: StockHolding[]): HoldingGroup[] {
	const groups: HoldingGroup[] = [];
	for (const holding of holdings) {
		const last = groups.at(-1);
		if (last?.productId === holding.product_id) {
			last.balance += holding.balance;
			last.holdings.push(holding);
			continue;
		}
		groups.push({
			productId: holding.product_id,
			productName: holding.product_name,
			unit: holding.unit,
			balance: holding.balance,
			holdings: [holding]
		});
	}
	return groups;
}

export function sourceHrefOf(sourceType: string, sourceId: string): string {
	return sourceType === 'DROP' ? `/app/titipan/${sourceId}` : `/app/hitung/${sourceId}`;
}
