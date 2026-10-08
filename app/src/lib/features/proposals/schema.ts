import type { Product } from '#lib/api/products.js';
import type { CountRhythm, InvitationRequest, SettlementSchedule } from '#lib/api/proposals.js';
import type { PricingModel, Terms } from '#lib/domain/pricing.js';
import { id } from '#lib/i18n/id.js';

const basisPointsPerPercent = 100;
const maxCommissionPercent = 50;
const defaultPaymentTermsDays = 7;
const defaultCommissionPercent = '20';

export interface ItemValues {
	productId: string;
	name: string;
	unit: string;
	model: PricingModel;
	retailPrice: number | undefined;
	depositPrice: number | undefined;
	commissionPercent: string;
}

export interface InvitationValues {
	schedule: SettlementSchedule;
	paymentTermsDays: number;
	rhythm: CountRhythm;
	notes: string;
	message: string;
	items: ItemValues[];
}

export type InvitationErrors = Record<string, string>;

export function emptyInvitation(): InvitationValues {
	return {
		schedule: 'WEEKLY',
		paymentTermsDays: defaultPaymentTermsDays,
		rhythm: 'DAILY',
		notes: '',
		message: '',
		items: []
	};
}

export interface CatalogEntry {
	id: string;
	name: string;
	unit: string;
	retailPrice: number;
	depositPrice: number | undefined;
}

export function entryOfProduct(product: Product): CatalogEntry {
	return {
		id: product.id,
		name: product.name,
		unit: product.unit,
		retailPrice: product.suggested_retail_price,
		depositPrice: product.suggested_deposit_price
	};
}

export function itemFromEntry(entry: CatalogEntry): ItemValues {
	return {
		productId: entry.id,
		name: entry.name,
		unit: entry.unit,
		model: entry.depositPrice === undefined ? 'COMMISSION' : 'FIXED_DEPOSIT',
		retailPrice: entry.retailPrice,
		depositPrice: entry.depositPrice,
		commissionPercent: defaultCommissionPercent
	};
}

export function commissionBps(percentText: string): number | undefined {
	const value = Number(percentText.trim().replace(',', '.'));
	if (percentText.trim() === '' || !Number.isFinite(value)) {
		return undefined;
	}
	return Math.round(value * basisPointsPerPercent);
}

export function termsOf(item: ItemValues): Terms {
	const isCommission = item.model === 'COMMISSION';
	return {
		model: item.model,
		retailPrice: item.retailPrice ?? 0,
		commissionBps: isCommission ? (commissionBps(item.commissionPercent) ?? 0) : 0,
		depositPrice: isCommission ? 0 : (item.depositPrice ?? 0)
	};
}

export function toInvitationRequest(values: InvitationValues): InvitationRequest {
	return {
		terms: {
			settlement_schedule: values.schedule,
			payment_terms_days: values.paymentTermsDays,
			count_rhythm: values.rhythm,
			notes: values.notes
		},
		items: values.items.map((item) => {
			const terms = termsOf(item);
			return {
				product_id: item.productId,
				pricing_model: terms.model,
				retail_price: terms.retailPrice,
				commission_bps: terms.commissionBps,
				deposit_price: terms.depositPrice
			};
		}),
		message: values.message
	};
}

function itemProblem(item: ItemValues): [string, string] | undefined {
	if (item.model === 'FIXED_DEPOSIT') {
		const isOrdered = (item.depositPrice ?? 0) <= (item.retailPrice ?? 0);
		return isOrdered ? undefined : ['deposit_price', id.products.priceOrder];
	}
	const bps = commissionBps(item.commissionPercent);
	const isInRange =
		bps !== undefined && bps >= 0 && bps <= maxCommissionPercent * basisPointsPerPercent;
	return isInRange ? undefined : ['commission_bps', id.proposals.invalidPercent];
}

export function validateInvitation(values: InvitationValues): InvitationErrors {
	const errors: InvitationErrors = {};
	values.items.forEach((item, index) => {
		const problem = itemProblem(item);
		if (problem) {
			errors[`items[${String(index)}].${problem[0]}`] = problem[1];
		}
	});
	return errors;
}
