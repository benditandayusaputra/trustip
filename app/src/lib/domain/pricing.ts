export type PricingModel = 'FIXED_DEPOSIT' | 'COMMISSION';

export interface Terms {
	model: PricingModel;
	retailPrice: number;
	commissionBps: number;
	depositPrice: number;
}

export interface Split {
	gross: number;
	storeShare: number;
	consignorShare: number;
}

export const exampleQuantity = 10;
const wholeBps = 10_000;
const halfBps = 5_000;

type ShareRule = (terms: Terms, quantity: number, gross: number) => Split;

const shareRules: Record<PricingModel, ShareRule> = {
	COMMISSION: (terms, _quantity, gross) => {
		const storeShare = Math.floor((gross * terms.commissionBps + halfBps) / wholeBps);
		return { gross, storeShare, consignorShare: gross - storeShare };
	},
	FIXED_DEPOSIT: (terms, quantity, gross) => {
		const consignorShare = quantity * terms.depositPrice;
		return { gross, storeShare: gross - consignorShare, consignorShare };
	}
};

export function splitOf(terms: Terms, quantity: number): Split {
	return shareRules[terms.model](terms, quantity, quantity * terms.retailPrice);
}

export function summarize(splits: Split[]): Split {
	return splits.reduce(
		(total, split) => ({
			gross: total.gross + split.gross,
			storeShare: total.storeShare + split.storeShare,
			consignorShare: total.consignorShare + split.consignorShare
		}),
		{ gross: 0, storeShare: 0, consignorShare: 0 }
	);
}
