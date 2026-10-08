import type { Agreement } from '#lib/api/agreements.js';
import type { DropRequest } from '#lib/api/drops.js';
import type { PartnerCatalogItem } from '#lib/api/partnerships.js';
import { addDays, daysBetween } from '#lib/domain/dates.js';
import { splitOf, summarize, type Split, type Terms } from '#lib/domain/pricing.js';
import { id } from '#lib/i18n/id.js';

export const maxQty = 10_000;
export const maxDaysBack = 3;
const warningDaysLeft = 2;
const warningMinShelfLife = 7;
const backdatedClock = 'T12:00:00+07:00';

export interface DropLine {
	productId: string;
	name: string;
	unit: string;
	qty: number;
	expiryDate: string;
	shelfLifeDays: number | null;
	terms: Terms;
}

export type DropErrors = Record<string, string>;

type AgreementItem = Agreement['items'][number];

export function defaultExpiry(dropDate: string, shelfLifeDays: number | null): string {
	return shelfLifeDays === null ? '' : addDays(dropDate, shelfLifeDays);
}

export function linesOf(
	agreement: Agreement,
	catalog: PartnerCatalogItem[],
	dropDate: string
): DropLine[] {
	const shelfLives = new Map(catalog.map((item) => [item.product_id, item.shelf_life_days]));
	return agreement.items
		.filter((item) => item.is_active)
		.map((item) => lineOf(item, shelfLives.get(item.product_id) ?? null, dropDate));
}

function lineOf(item: AgreementItem, shelfLifeDays: number | null, dropDate: string): DropLine {
	return {
		productId: item.product_id,
		name: item.product_name,
		unit: item.unit,
		qty: 0,
		expiryDate: defaultExpiry(dropDate, shelfLifeDays),
		shelfLifeDays,
		terms: {
			model: item.pricing_model,
			retailPrice: item.retail_price,
			commissionBps: item.commission_bps,
			depositPrice: item.deposit_price
		}
	};
}

export function filledLines(lines: DropLine[]): DropLine[] {
	return lines.filter((line) => line.qty > 0);
}

export function isExpirySoon(line: DropLine, dropDate: string): boolean {
	if (line.expiryDate === '' || line.shelfLifeDays === null) {
		return false;
	}
	const daysLeft = daysBetween(dropDate, line.expiryDate);
	return daysLeft >= 0 && daysLeft < warningDaysLeft && line.shelfLifeDays > warningMinShelfLife;
}

export function validateDrop(lines: DropLine[], dropDate: string): DropErrors {
	const errors: DropErrors = {};
	if (filledLines(lines).length === 0) {
		errors.items = id.drops.nothingFilled;
	}
	for (const line of filledLines(lines)) {
		if (line.expiryDate !== '' && line.expiryDate < dropDate) {
			errors[`${line.productId}.expiry`] = id.drops.expiryBeforeDrop;
		}
	}
	return errors;
}

export interface DropPreview {
	units: number;
	split: Split;
}

export function previewOf(lines: DropLine[]): DropPreview {
	const filled = filledLines(lines);
	return {
		units: filled.reduce((total, line) => total + line.qty, 0),
		split: summarize(filled.map((line) => splitOf(line.terms, line.qty)))
	};
}

export interface DropDraft {
	partnershipId: string;
	agreementVersion: number;
	dropDate: string;
	today: string;
	lines: DropLine[];
}

export function toDropRequest(draft: DropDraft): DropRequest {
	return {
		partnership_id: draft.partnershipId,
		agreement_version: draft.agreementVersion,
		occurred_at: draft.dropDate === draft.today ? null : `${draft.dropDate}${backdatedClock}`,
		items: filledLines(draft.lines).map((line) => ({
			product_id: line.productId,
			qty: line.qty,
			expiry_date: line.expiryDate === '' ? null : line.expiryDate
		}))
	};
}

const itemField = /^items\[(\d+)\]\.(\w+)$/u;
const fieldNames: Record<string, string> = { qty: 'qty', expiry_date: 'expiry', product_id: 'qty' };

export function errorKeyOf(field: string, sent: DropLine[]): string {
	const match = itemField.exec(field);
	const line = match ? sent[Number(match[1])] : undefined;
	if (!match || !line) {
		return field === 'occurred_at' ? 'dropDate' : 'items';
	}
	return `${line.productId}.${fieldNames[match[2] ?? ''] ?? 'qty'}`;
}
