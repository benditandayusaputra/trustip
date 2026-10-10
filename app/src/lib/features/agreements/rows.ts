import type { Agreement } from '#lib/api/agreements.js';
import { formatBasisPoints, formatRupiah } from '#lib/i18n/format.js';
import { agreements as agreementsText } from '#lib/i18n/agreements.js';
import { proposals as proposalsText } from '#lib/i18n/proposals.js';

const id = { ...agreementsText, ...proposalsText };

export interface ProductRow extends Record<string, string> {
	id: string;
	product: string;
	model: string;
	retail: string;
	consignor: string;
	store: string;
}

export function productRows(agreement: Agreement): ProductRow[] {
	const models = id.proposals.models;
	return agreement.items.map((item) => ({
		id: item.product_id,
		product: `${item.product_name} (${item.sku})${item.is_active ? '' : ` · ${id.agreement.stopped}`}`,
		model:
			item.pricing_model === 'COMMISSION'
				? `${models.COMMISSION} ${formatBasisPoints(item.commission_bps)}`
				: models.FIXED_DEPOSIT,
		retail: formatRupiah(item.retail_price),
		consignor: formatRupiah(item.per_unit.consignor_share),
		store: formatRupiah(item.per_unit.store_share)
	}));
}
