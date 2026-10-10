import type { Agreement } from '#lib/api/agreements.js';
import type { CountSheet } from '#lib/api/counts.js';
import { formatBusinessDate, formatInteger } from '#lib/i18n/format.js';
import { drops as dropsText } from '#lib/i18n/drops.js';
import { offline as offlineText } from '#lib/i18n/offline.js';
import type { OutboxEntry } from '#lib/offline/outbox.js';

const id = { ...dropsText, ...offlineText };

type CountEntry = Extract<OutboxEntry, { kind: 'COUNT' }>;
type DropEntry = Extract<OutboxEntry, { kind: 'DROP' }>;

export type CountReviewRow = Record<
	'productId' | 'product' | 'seen' | 'latest' | 'counted' | 'returned',
	string
>;

export type DropReviewRow = Record<'productId' | 'product' | 'qty' | 'expiry' | 'agreed', string>;

export function countReviewRows(entry: CountEntry, sheet: CountSheet): CountReviewRow[] {
	return entry.request.lines.map((line) => {
		const latest = sheet.products.find((product) => product.product_id === line.product_id);
		return {
			productId: line.product_id,
			product: latest?.product_name ?? id.outbox.noStockNow,
			seen: formatInteger(line.recorded_qty),
			latest: formatInteger(latest?.recorded_qty ?? 0),
			counted: formatInteger(line.counted_qty),
			returned: formatInteger(line.returned_qty)
		};
	});
}

export function dropReviewRows(entry: DropEntry, agreement: Agreement): DropReviewRow[] {
	const agreed = new Set(
		agreement.items.filter((item) => item.is_active).map((item) => item.product_id)
	);
	return entry.lines.map((line) => ({
		productId: line.productId,
		product: line.name,
		qty: `${formatInteger(line.qty)} ${line.unit}`,
		expiry: line.expiryDate === '' ? id.drops.noExpiry : formatBusinessDate(line.expiryDate),
		agreed: agreed.has(line.productId) ? id.outbox.agreedYes : id.outbox.agreedNo
	}));
}

export function fixHrefOf(entry: OutboxEntry): string {
	const forms = { DROP: '/app/titipan/baru', COUNT: '/app/hitung/baru' };
	const query = new URLSearchParams({ mitra: entry.partnershipId, antrean: entry.id });
	return `${forms[entry.kind]}?${query.toString()}`;
}
