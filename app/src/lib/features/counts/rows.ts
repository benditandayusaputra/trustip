import type { Count } from '#lib/api/counts.js';
import { formatInteger, formatRupiah } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';

export interface CountLineRow extends Record<string, string> {
	id: string;
	product: string;
	recorded: string;
	sold: string;
	returned: string;
	kept: string;
	gross: string;
}

export function lineRows(count: Count): CountLineRow[] {
	const text = id.counts;
	return count.lines.map((line) => ({
		id: line.id,
		product: `${line.product_name} (${line.sku})`,
		recorded: formatInteger(line.recorded_qty),
		sold: formatInteger(line.sold_qty),
		returned:
			line.return_reason === null
				? formatInteger(line.returned_qty)
				: `${formatInteger(line.returned_qty)} · ${text.reasons[line.return_reason]}`,
		kept: formatInteger(line.kept_qty),
		gross: formatRupiah(line.split.gross)
	}));
}
