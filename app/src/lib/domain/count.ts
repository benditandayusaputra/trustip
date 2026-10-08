import { splitOf, summarize, type Split, type Terms } from './pricing';
import { allocate, deduct, type Allocation, type Batch } from './stock';

export type ReturnReason = 'UNSOLD' | 'EXPIRED' | 'NEAR_EXPIRY' | 'DAMAGED' | 'RECALL' | 'OTHER';

interface PricedBatch extends Batch {
	terms: Terms;
}

export interface LineInput {
	productId: string;
	batches: PricedBatch[];
	countedQty: number;
	returnedQty: number;
	returnReason: ReturnReason | '';
}

interface SaleAllocation extends Allocation {
	split: Split;
}

interface Line {
	productId: string;
	recordedQty: number;
	countedQty: number;
	soldQty: number;
	returnedQty: number;
	keptQty: number;
	returnReason: ReturnReason | '';
	sales: SaleAllocation[];
	returns: Allocation[];
	split: Split;
}

export interface Tally {
	lines: Line[];
	soldQty: number;
	returnedQty: number;
	split: Split;
}

export type LineProblem = 'COUNT_EXCEEDS_STOCK' | 'RETURNED_OUT_OF_RANGE' | 'REASON_REQUIRED';

function recordedQty(batches: Batch[]): number {
	return batches.reduce((total, batch) => total + batch.balance, 0);
}

export function findLineProblem(input: LineInput): LineProblem | null {
	if (input.countedQty < 0 || input.countedQty > recordedQty(input.batches)) {
		return 'COUNT_EXCEEDS_STOCK';
	}
	if (input.returnedQty < 0 || input.returnedQty > input.countedQty) {
		return 'RETURNED_OUT_OF_RANGE';
	}
	if (input.returnedQty > 0 && input.returnReason === '') {
		return 'REASON_REQUIRED';
	}
	return null;
}

function tallyLine(input: LineInput): Line {
	const recorded = recordedQty(input.batches);
	const soldQty = recorded - input.countedQty;
	const sold = allocate(input.batches, soldQty);
	const returns = allocate(deduct(input.batches, sold), input.returnedQty);
	const termsByBatch = new Map(input.batches.map((batch) => [batch.id, batch.terms]));
	const sales = sold.map((allocation) => {
		const terms = termsByBatch.get(allocation.batchId);
		if (terms === undefined) {
			throw new Error(`missing terms for batch ${allocation.batchId}`);
		}
		return { ...allocation, split: splitOf(terms, allocation.qty) };
	});
	return {
		productId: input.productId,
		recordedQty: recorded,
		countedQty: input.countedQty,
		soldQty,
		returnedQty: input.returnedQty,
		keptQty: input.countedQty - input.returnedQty,
		returnReason: input.returnedQty > 0 ? input.returnReason : '',
		sales,
		returns,
		split: summarize(sales.map((sale) => sale.split))
	};
}

export function tallyCount(inputs: LineInput[]): Tally {
	const lines = inputs.map(tallyLine);
	return {
		lines,
		soldQty: lines.reduce((total, line) => total + line.soldQty, 0),
		returnedQty: lines.reduce((total, line) => total + line.returnedQty, 0),
		split: summarize(lines.map((line) => line.split))
	};
}
