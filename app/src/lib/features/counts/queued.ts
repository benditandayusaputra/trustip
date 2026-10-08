import type { CountSheet } from '#lib/api/counts.js';
import type { OutboxEntry, QueuedDropLine } from '#lib/offline/outbox.js';

type SheetProduct = CountSheet['products'][number];
type SheetBatch = SheetProduct['batches'][number];
type QueuedDrop = Extract<OutboxEntry, { kind: 'DROP' }>;

function isQueuedDropFor(partnershipId: string) {
	return (entry: OutboxEntry): entry is QueuedDrop =>
		entry.kind === 'DROP' && entry.partnershipId === partnershipId && entry.status !== 'SENT';
}

export function withQueuedDrops(sheet: CountSheet, entries: OutboxEntry[]): CountSheet {
	return entries.filter(isQueuedDropFor(sheet.partnership_id)).reduce(addDrop, sheet);
}

function addDrop(sheet: CountSheet, drop: QueuedDrop): CountSheet {
	let products = sheet.products;
	for (const [index, line] of drop.lines.entries()) {
		products = addLine(products, line, {
			id: `${drop.id}/${String(index + 1)}`,
			at: drop.recordedAt
		});
	}
	const lastDocumentAt =
		sheet.last_document_at !== null && sheet.last_document_at > drop.recordedAt
			? sheet.last_document_at
			: drop.recordedAt;
	return { ...sheet, products, last_document_at: lastDocumentAt };
}

interface BatchSource {
	id: string;
	at: string;
}

function batchOf(line: QueuedDropLine, source: BatchSource): SheetBatch {
	return {
		id: source.id,
		expiry_date: line.expiryDate === '' ? null : line.expiryDate,
		received_at: source.at,
		balance: line.qty,
		terms: {
			pricing_model: line.terms.model,
			retail_price: line.terms.retailPrice,
			commission_bps: line.terms.commissionBps,
			deposit_price: line.terms.depositPrice
		}
	};
}

function addLine(
	products: SheetProduct[],
	line: QueuedDropLine,
	source: BatchSource
): SheetProduct[] {
	const batch = batchOf(line, source);
	const current = products.find((product) => product.product_id === line.productId);
	if (!current) {
		const product = {
			product_id: line.productId,
			product_name: line.name,
			sku: '',
			unit: line.unit
		};
		return [...products, { ...product, recorded_qty: line.qty, batches: [batch] }];
	}
	const grown = {
		...current,
		recorded_qty: current.recorded_qty + line.qty,
		batches: [...current.batches, batch]
	};
	return products.map((product) => (product === current ? grown : product));
}
