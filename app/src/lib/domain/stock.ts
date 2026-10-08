export type BatchFlag = 'SAFE' | 'NEAR_EXPIRY' | 'EXPIRED';

export interface Movement {
	batchId: string;
	qty: number;
}

export interface Batch {
	id: string;
	expiryDate: string | null;
	receivedAt: string;
	balance: number;
}

export interface Allocation {
	batchId: string;
	qty: number;
}

export class InsufficientStockError extends Error {
	constructor(qty: number) {
		super(`cannot allocate ${String(qty)} units from the given batches`);
		this.name = 'InsufficientStockError';
	}
}

const millisecondsPerDay = 86_400_000;
const nearExpiryDays = 7;
const shortShelfLifeDays = 7;
const shortShelfLifeNearDays = 1;

export function balances(movements: Movement[]): Map<string, number> {
	const totals = new Map<string, number>();
	for (const movement of movements) {
		totals.set(movement.batchId, (totals.get(movement.batchId) ?? 0) + movement.qty);
	}
	return totals;
}

export function allocate(batches: Batch[], qty: number): Allocation[] {
	const total = batches.reduce((sum, batch) => sum + Math.max(batch.balance, 0), 0);
	if (qty < 0 || qty > total) {
		throw new InsufficientStockError(qty);
	}
	const allocations: Allocation[] = [];
	let unallocated = qty;
	for (const batch of [...batches].sort(compareLeavingOrder)) {
		const taken = Math.min(unallocated, batch.balance);
		if (taken > 0) {
			allocations.push({ batchId: batch.id, qty: taken });
			unallocated -= taken;
		}
	}
	return allocations;
}

export function deduct(batches: Batch[], allocations: Allocation[]): Batch[] {
	const taken = balances(allocations);
	return batches.map((batch) => ({
		...batch,
		balance: batch.balance - (taken.get(batch.id) ?? 0)
	}));
}

function compareLeavingOrder(left: Batch, right: Batch): number {
	return (
		compareExpiry(left.expiryDate, right.expiryDate) ||
		Date.parse(left.receivedAt) - Date.parse(right.receivedAt) ||
		compareText(left.id, right.id)
	);
}

function compareExpiry(left: string | null, right: string | null): number {
	if (left === null || right === null) {
		return Number(left === null) - Number(right === null);
	}
	return compareText(left, right);
}

function compareText(left: string, right: string): number {
	if (left === right) {
		return 0;
	}
	return left < right ? -1 : 1;
}

export function flagOn(
	expiryDate: string | null,
	today: string,
	shelfLifeDays: number | null
): BatchFlag {
	if (expiryDate === null) {
		return 'SAFE';
	}
	const daysLeft = (Date.parse(expiryDate) - Date.parse(today)) / millisecondsPerDay;
	if (daysLeft < 0) {
		return 'EXPIRED';
	}
	const isShortShelfLife = shelfLifeDays !== null && shelfLifeDays <= shortShelfLifeDays;
	const threshold = isShortShelfLife ? shortShelfLifeNearDays : nearExpiryDays;
	return daysLeft <= threshold ? 'NEAR_EXPIRY' : 'SAFE';
}
