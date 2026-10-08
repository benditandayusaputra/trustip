import type { Settlement, SettlementSummary } from '#lib/api/settlements.js';
import type { Tone } from '#lib/components/ui/tone.js';
import { formatBasisPoints, formatDate, formatInteger, formatRupiah } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';

export interface SettlementLineRow extends Record<string, string> {
	key: string;
	product: string;
	model: string;
	qty: string;
	retail: string;
	gross: string;
	storeShare: string;
	consignorShare: string;
}

const lastMomentBeforeEnd = 1000;

export function lineRows(settlement: Settlement): SettlementLineRow[] {
	const text = id.settlements;
	return settlement.lines.map((line, index) => ({
		key: `${line.product_id}-${String(index)}`,
		product: `${line.product_name} (${line.sku})`,
		model:
			line.pricing_model === 'COMMISSION'
				? `${text.models.COMMISSION} ${formatBasisPoints(line.commission_bps)}`
				: `${text.models.FIXED_DEPOSIT} ${formatRupiah(line.deposit_price)}`,
		qty: formatInteger(line.qty),
		retail: formatRupiah(line.retail_price),
		gross: formatRupiah(line.split.gross),
		storeShare: formatRupiah(line.split.store_share),
		consignorShare: formatRupiah(line.split.consignor_share)
	}));
}

export function periodLabel(period: Settlement['period']): string {
	const lastMoment = new Date(Date.parse(period.end) - lastMomentBeforeEnd).toISOString();
	return id.settlements.periodRange(formatDate(period.start), formatDate(lastMoment));
}

type Standing = SettlementSummary['standing'];

export function standingTone(standing: Standing): Tone {
	if (standing.age === 'PAID') {
		return 'success';
	}
	return standing.is_overdue ? 'danger' : 'info';
}

export function standingLabel(standing: Standing): string {
	const text = id.settlements;
	if (standing.age === 'PAID') {
		return text.statuses.PAID;
	}
	return standing.is_overdue ? text.daysLate(standing.days_late) : text.statuses.ISSUED;
}
