import type { CountOccasion, CountRequest, CountSheet } from '#lib/api/counts.js';
import {
	findLineProblem,
	tallyCount,
	type LineInput,
	type ReturnReason,
	type Tally
} from '#lib/domain/count.js';
import { addDays } from '#lib/domain/dates.js';
import { businessDateOf } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';

const backdatedClock = 'T12:00:00+07:00';
const maxDaysBack = 3;
export const maxNoteLength = 500;

export function earliestCountDate(today: string, lastDocumentAt: string | null): string {
	const earliest = addDays(today, -maxDaysBack);
	const lastDay = lastDocumentAt === null ? '' : businessDateOf(new Date(lastDocumentAt));
	return lastDay > earliest ? lastDay : earliest;
}

export type InputMode = 'remaining' | 'sold';

export interface CountLineValues {
	productId: string;
	name: string;
	unit: string;
	recordedQty: number;
	countedQty: number;
	returnedQty: number;
	returnReason: ReturnReason | '';
	input: LineInput;
}

export type CountErrors = Record<string, string>;

type SheetProduct = CountSheet['products'][number];

function batchesOf(product: SheetProduct): LineInput['batches'] {
	return product.batches.map((batch) => ({
		id: batch.id,
		expiryDate: batch.expiry_date,
		receivedAt: batch.received_at,
		balance: batch.balance,
		terms: {
			model: batch.terms.pricing_model,
			retailPrice: batch.terms.retail_price,
			commissionBps: batch.terms.commission_bps,
			depositPrice: batch.terms.deposit_price
		}
	}));
}

export function linesOf(sheet: CountSheet): CountLineValues[] {
	return sheet.products.map((product) => ({
		productId: product.product_id,
		name: product.product_name,
		unit: product.unit,
		recordedQty: product.recorded_qty,
		countedQty: product.recorded_qty,
		returnedQty: 0,
		returnReason: '',
		input: {
			productId: product.product_id,
			batches: batchesOf(product),
			countedQty: product.recorded_qty,
			returnedQty: 0,
			returnReason: ''
		}
	}));
}

export function withChanges(
	line: CountLineValues,
	changes: Partial<Pick<CountLineValues, 'countedQty' | 'returnedQty' | 'returnReason'>>
): CountLineValues {
	const next = { ...line, ...changes };
	return {
		...next,
		input: {
			...line.input,
			countedQty: next.countedQty,
			returnedQty: next.returnedQty,
			returnReason: next.returnReason
		}
	};
}

export interface ExpiredReturn {
	productId: string;
	qty: number;
}

export function expiredReturnOf(params: Pick<URLSearchParams, 'get'>): ExpiredReturn | undefined {
	const productId = params.get('produk') ?? '';
	const qty = Number(params.get('kembali'));
	return productId !== '' && Number.isInteger(qty) && qty > 0 ? { productId, qty } : undefined;
}

export function withExpiredReturn(
	lines: CountLineValues[],
	expired: ExpiredReturn
): CountLineValues[] {
	return lines.map((line) =>
		line.productId === expired.productId
			? withChanges(line, {
					returnedQty: Math.min(expired.qty, line.countedQty),
					returnReason: 'EXPIRED'
				})
			: line
	);
}

export function allSold(lines: CountLineValues[]): CountLineValues[] {
	return lines.map((line) =>
		withChanges(line, { countedQty: 0, returnedQty: 0, returnReason: '' })
	);
}

export function returnAllRemaining(lines: CountLineValues[]): CountLineValues[] {
	return lines.map((line) =>
		withChanges(line, {
			returnedQty: line.countedQty,
			returnReason: line.countedQty > 0 ? line.returnReason || 'UNSOLD' : ''
		})
	);
}

function lineErrorOf(line: CountLineValues): string | undefined {
	const messages = {
		COUNT_EXCEEDS_STOCK: id.counts.exceedsStock(line.recordedQty),
		RETURNED_OUT_OF_RANGE: id.counts.returnedTooMany,
		REASON_REQUIRED: id.counts.reasonRequired
	};
	const problem = findLineProblem(line.input);
	return problem === null ? undefined : messages[problem];
}

export function validateCount(lines: CountLineValues[]): CountErrors {
	const errors: CountErrors = {};
	for (const line of lines) {
		const message = lineErrorOf(line);
		if (message !== undefined) {
			errors[`${line.productId}.line`] = message;
		}
	}
	return errors;
}

export function previewOf(lines: CountLineValues[]): Tally {
	return tallyCount(
		lines.filter((line) => lineErrorOf(line) === undefined).map((line) => line.input)
	);
}

export interface CountDraft {
	partnershipId: string;
	agreementVersion: number;
	countedOn: string;
	today: string;
	occasion: CountOccasion;
	note: string;
	lines: CountLineValues[];
}

export function toCountRequest(draft: CountDraft): CountRequest {
	return {
		partnership_id: draft.partnershipId,
		agreement_version: draft.agreementVersion,
		counted_at: draft.countedOn === draft.today ? null : `${draft.countedOn}${backdatedClock}`,
		occasion: draft.occasion,
		note: draft.note,
		lines: draft.lines.map((line) => ({
			product_id: line.productId,
			recorded_qty: line.recordedQty,
			counted_qty: line.countedQty,
			returned_qty: line.returnedQty,
			return_reason: line.returnReason === '' ? null : line.returnReason
		}))
	};
}

const lineField = /^lines\[(\d+)\]\.\w+$/u;

export function errorKeyOf(field: string, lines: CountLineValues[]): string {
	const match = lineField.exec(field);
	const line = match ? lines[Number(match[1])] : undefined;
	if (line) {
		return `${line.productId}.line`;
	}
	return field === 'counted_at' ? 'countedOn' : field;
}

export function revisionDateOf(countedAt: string | null, recordedAt: string): string {
	return businessDateOf(new Date(countedAt ?? recordedAt));
}

export function refilledCountLines(
	lines: CountLineValues[],
	sent: CountRequest['lines']
): CountLineValues[] {
	return lines.map((line) => {
		const previous = sent.find((value) => value.product_id === line.productId);
		if (!previous) {
			return line;
		}
		return withChanges(line, {
			countedQty: previous.counted_qty,
			returnedQty: previous.returned_qty,
			returnReason: previous.return_reason ?? ''
		});
	});
}
