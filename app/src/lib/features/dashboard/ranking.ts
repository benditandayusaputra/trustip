import type { SalesReport } from '#lib/api/reports.js';

export interface RankedRow {
	key: string;
	label: string;
	value: string;
	product?: SalesReport['rows'][number]['product'];
}
