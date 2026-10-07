import type { AgreementVersion } from '#lib/api/agreements.js';
import { formatBasisPoints, formatRupiah } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';

type TermChange = AgreementVersion['term_changes'][number];
type ItemChange = AgreementVersion['item_changes'][number];
type Snapshot = ItemChange['after'];

const terms = id.proposals;

const labels: Record<TermChange['field'], string> = {
	settlement_schedule: terms.schedule,
	payment_terms_days: terms.paymentTerms,
	count_rhythm: terms.rhythm,
	notes: terms.notes
};

function valueOf(field: TermChange['field'], raw: string): string {
	const choices: Partial<Record<TermChange['field'], Record<string, string>>> = {
		settlement_schedule: terms.schedules,
		count_rhythm: terms.rhythms
	};
	return choices[field]?.[raw] ?? (raw === '' ? id.agreement.emptyNotes : raw);
}

export function describeTermChange(change: TermChange): string {
	return id.agreement.termChange(
		labels[change.field],
		valueOf(change.field, change.before),
		valueOf(change.field, change.after)
	);
}

function describeTerms(snapshot: Snapshot): string {
	const price = formatRupiah(snapshot.retail_price);
	if (snapshot.pricing_model === 'COMMISSION') {
		return `${terms.models.COMMISSION} ${formatBasisPoints(snapshot.commission_bps)}, ${price}`;
	}
	return `${terms.models.FIXED_DEPOSIT} ${formatRupiah(snapshot.deposit_price)}, ${price}`;
}

export function describeItemChange(change: ItemChange): string {
	const kind = id.agreement.changeKinds[change.kind];
	if (change.kind === 'STOPPED') {
		return `${kind}: ${change.product_name}`;
	}
	const after = describeTerms(change.after);
	const before = change.before ? `${describeTerms(change.before)} → ` : '';
	return `${kind}: ${change.product_name} (${before}${after})`;
}
