import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import {
	allocate,
	balances,
	deduct,
	flagOn,
	InsufficientStockError,
	type Allocation,
	type Batch
} from './stock';

interface VectorBatch {
	id: string;
	expiry_date: string | null;
	received_at: string;
	balance: number;
}

interface VectorAllocation {
	batch_id: string;
	qty: number;
}

interface Vectors {
	balances: {
		name: string;
		movements: VectorAllocation[];
		balances: Record<string, number>;
	}[];
	allocations: {
		name: string;
		batches: VectorBatch[];
		qty: number;
		allocations: VectorAllocation[];
		remaining: Record<string, number>;
	}[];
	shortages: { name: string; batches: VectorBatch[]; qty: number }[];
	flags: {
		name: string;
		expiry_date: string | null;
		shelf_life_days: number | null;
		today: string;
		flag: string;
	}[];
}

const vectors = JSON.parse(
	readFileSync(new URL('../../../../api/testdata/vectors/stock.json', import.meta.url), 'utf8')
) as Vectors;

const batchesOf = (raw: VectorBatch[]): Batch[] =>
	raw.map((batch) => ({
		id: batch.id,
		expiryDate: batch.expiry_date,
		receivedAt: batch.received_at,
		balance: batch.balance
	}));

const allocationsOf = (raw: VectorAllocation[]): Allocation[] =>
	raw.map((allocation) => ({ batchId: allocation.batch_id, qty: allocation.qty }));

const balancesOf = (batches: Batch[]) =>
	Object.fromEntries(batches.map((batch) => [batch.id, batch.balance]));

describe('stock against the shared vectors', () => {
	it.each(vectors.balances)('sums movements per batch: $name', (vector) => {
		expect(Object.fromEntries(balances(allocationsOf(vector.movements)))).toEqual(vector.balances);
	});

	it.each(vectors.allocations)('BR-STK-02 $name', (vector) => {
		const batches = batchesOf(vector.batches);

		const allocations = allocate(batches, vector.qty);

		expect(allocations).toEqual(allocationsOf(vector.allocations));
		expect(balancesOf(deduct(batches, allocations))).toEqual(vector.remaining);
	});

	it.each(vectors.shortages)('BR-STK-01 $name', (vector) => {
		expect(() => allocate(batchesOf(vector.batches), vector.qty)).toThrow(InsufficientStockError);
	});

	it.each(vectors.flags)('BR-STK-04 $name', (vector) => {
		expect(flagOn(vector.expiry_date, vector.today, vector.shelf_life_days)).toBe(vector.flag);
	});
});
