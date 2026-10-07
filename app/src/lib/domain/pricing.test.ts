import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { exampleQuantity, splitOf, summarize, type PricingModel, type Split } from './pricing';

interface VectorSplit {
	gross: number;
	store_share: number;
	consignor_share: number;
}

interface Allocation {
	name: string;
	terms: {
		model: PricingModel;
		retail_price: number;
		commission_bps: number;
		deposit_price: number;
	};
	qty: number;
	split: VectorSplit;
}

interface Vectors {
	allocations: Allocation[];
	summary: { lines: number[]; total: VectorSplit };
	example_quantity: number;
}

const vectors = JSON.parse(
	readFileSync(new URL('../../../../api/testdata/vectors/pricing.json', import.meta.url), 'utf8')
) as Vectors;

const termsOf = (allocation: Allocation) => ({
	model: allocation.terms.model,
	retailPrice: allocation.terms.retail_price,
	commissionBps: allocation.terms.commission_bps,
	depositPrice: allocation.terms.deposit_price
});

const asVector = (split: Split): VectorSplit => ({
	gross: split.gross,
	store_share: split.storeShare,
	consignor_share: split.consignorShare
});

describe('pricing against the shared vectors', () => {
	it.each(vectors.allocations)('BR-STL-04 $name', (allocation) => {
		expect(asVector(splitOf(termsOf(allocation), allocation.qty))).toEqual(allocation.split);
	});

	it('BR-STL-04 totals add the lines without recomputing commission', () => {
		const splits = vectors.summary.lines.map((index) => {
			const allocation = vectors.allocations[index];
			if (!allocation) {
				throw new Error(`missing allocation ${String(index)}`);
			}
			return splitOf(termsOf(allocation), allocation.qty);
		});

		expect(asVector(summarize(splits))).toEqual(vectors.summary.total);
	});

	it('FR-PRP-09 previews the same example quantity as the server', () => {
		expect(exampleQuantity).toBe(vectors.example_quantity);
	});
});
