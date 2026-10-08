import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import {
	findLineProblem,
	tallyCount,
	type LineInput,
	type LineProblem,
	type ReturnReason,
	type Tally
} from './count';
import type { PricingModel, Split } from './pricing';

interface VectorSplit {
	gross: number;
	store_share: number;
	consignor_share: number;
}

interface VectorLineInput {
	product_id: string;
	batches: {
		id: string;
		expiry_date: string | null;
		received_at: string;
		balance: number;
		terms: {
			model: PricingModel;
			retail_price: number;
			commission_bps: number;
			deposit_price: number;
		};
	}[];
	counted_qty: number;
	returned_qty: number;
	return_reason: ReturnReason | '';
}

interface VectorLine {
	product_id: string;
	recorded_qty: number;
	sold_qty: number;
	returned_qty: number;
	kept_qty: number;
	return_reason: string;
	sales: { batch_id: string; qty: number; split: VectorSplit }[];
	returns: { batch_id: string; qty: number }[];
	split: VectorSplit;
}

interface Vectors {
	counts: {
		name: string;
		lines: VectorLineInput[];
		result: { lines: VectorLine[]; sold_qty: number; returned_qty: number; split: VectorSplit };
	}[];
	rejections: {
		name: string;
		line: {
			recorded_qty: number;
			counted_qty: number;
			returned_qty: number;
			return_reason: string;
		};
		field: string;
		code: string;
	}[];
}

const vectors = JSON.parse(
	readFileSync(new URL('../../../../api/testdata/vectors/count.json', import.meta.url), 'utf8')
) as Vectors;

const inputOf = (raw: VectorLineInput): LineInput => ({
	productId: raw.product_id,
	batches: raw.batches.map((batch) => ({
		id: batch.id,
		expiryDate: batch.expiry_date,
		receivedAt: batch.received_at,
		balance: batch.balance,
		terms: {
			model: batch.terms.model,
			retailPrice: batch.terms.retail_price,
			commissionBps: batch.terms.commission_bps,
			depositPrice: batch.terms.deposit_price
		}
	})),
	countedQty: raw.counted_qty,
	returnedQty: raw.returned_qty,
	returnReason: raw.return_reason
});

const splitVector = (split: Split): VectorSplit => ({
	gross: split.gross,
	store_share: split.storeShare,
	consignor_share: split.consignorShare
});

const tallyVector = (tally: Tally) => ({
	lines: tally.lines.map((line) => ({
		product_id: line.productId,
		recorded_qty: line.recordedQty,
		sold_qty: line.soldQty,
		returned_qty: line.returnedQty,
		kept_qty: line.keptQty,
		return_reason: line.returnReason,
		sales: line.sales.map((sale) => ({
			batch_id: sale.batchId,
			qty: sale.qty,
			split: splitVector(sale.split)
		})),
		returns: line.returns.map((item) => ({ batch_id: item.batchId, qty: item.qty })),
		split: splitVector(line.split)
	})),
	sold_qty: tally.soldQty,
	returned_qty: tally.returnedQty,
	split: splitVector(tally.split)
});

const rejectionFields: Record<LineProblem, string> = {
	COUNT_EXCEEDS_STOCK: 'counted_qty',
	RETURNED_OUT_OF_RANGE: 'returned_qty',
	REASON_REQUIRED: 'return_reason'
};

describe('count against the shared vectors', () => {
	it.each(vectors.counts)('BR-CNT-08 $name', (vector) => {
		expect(tallyVector(tallyCount(vector.lines.map(inputOf)))).toEqual(vector.result);
	});

	it.each(vectors.rejections.filter((vector) => vector.code !== 'UNKNOWN_CHOICE'))(
		'$name',
		(vector) => {
			const problem = findLineProblem({
				productId: 'roti',
				batches: [
					{
						id: 'roti-1',
						expiryDate: null,
						receivedAt: '2026-11-01T00:00:00Z',
						balance: vector.line.recorded_qty,
						terms: { model: 'COMMISSION', retailPrice: 0, commissionBps: 0, depositPrice: 0 }
					}
				],
				countedQty: vector.line.counted_qty,
				returnedQty: vector.line.returned_qty,
				returnReason: vector.line.return_reason as ReturnReason | ''
			});

			expect(problem === null ? null : `lines[0].${rejectionFields[problem]}`).toBe(vector.field);
		}
	);
});
