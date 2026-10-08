import type { Drop } from '#lib/api/drops.js';
import { formatBusinessDate, formatInteger, formatRupiah } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';

export interface DropItemRow extends Record<string, string> {
	id: string;
	product: string;
	batch: string;
	expiry: string;
	qty: string;
	retail: string;
}

export function itemRows(drop: Drop): DropItemRow[] {
	return drop.items.map((item) => ({
		id: item.id,
		product: `${item.product_name} (${item.sku})`,
		batch: item.batch_code,
		expiry: item.expiry_date === null ? id.drops.noExpiry : formatBusinessDate(item.expiry_date),
		qty: `${formatInteger(item.qty)} ${item.unit}`,
		retail: formatRupiah(item.terms.retail_price)
	}));
}

export type StatusTone = 'pending' | 'success' | 'neutral';

export function statusToneOf(status: Drop['status']): StatusTone {
	const tones: Record<Drop['status'], StatusTone> = {
		RECORDED: 'pending',
		CONFIRMED: 'success',
		CANCELLED: 'neutral'
	};
	return tones[status];
}
