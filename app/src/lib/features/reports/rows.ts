import type { DailyRecap, SalesReport, SettlementReport } from '#lib/api/reports.js';
import { formatBasisPoints, formatInteger, formatRupiah } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';

type Figures = SalesReport['summary'];

export interface ReportRow extends Record<string, string> {
	key: string;
	label: string;
	sold: string;
	gross: string;
	storeShare: string;
	consignorShare: string;
	returned: string;
	returnedPercent: string;
}

function figureCells(figures: Figures) {
	return {
		sold: formatInteger(figures.sold_qty),
		gross: formatRupiah(figures.split.gross),
		storeShare: formatRupiah(figures.split.store_share),
		consignorShare: formatRupiah(figures.split.consignor_share),
		returned: formatInteger(figures.returned_qty),
		returnedPercent: formatBasisPoints(figures.returned_bps)
	};
}

export function reportRows(report: SalesReport): ReportRow[] {
	return report.rows.map((row) => ({ key: row.key, label: row.label, ...figureCells(row) }));
}

export interface SummaryCard {
	label: string;
	value: string;
}

export function summaryCards(summary: Figures, isStore: boolean): SummaryCard[] {
	const text = id.reports;
	const cells = figureCells(summary);
	return [
		{ label: text.soldUnits, value: cells.sold },
		{ label: text.gross, value: cells.gross },
		{ label: text.storeShare, value: cells.storeShare },
		{ label: isStore ? text.mustPay : text.consignorShare, value: cells.consignorShare },
		{ label: text.returnedUnits, value: cells.returned },
		{ label: text.returnedPercent, value: cells.returnedPercent }
	];
}

export interface RecapCells extends Record<string, string> {
	key: string;
	name: string;
	received: string;
	sold: string;
	returned: string;
	gross: string;
	storeShare: string;
	consignorShare: string;
	paid: string;
}

type RecapRow = DailyRecap['total'];

export function recapCells(row: RecapRow): RecapCells {
	return {
		key: row.partner_id,
		name: row.name,
		received: formatInteger(row.received_qty),
		sold: formatInteger(row.sold_qty),
		returned: formatInteger(row.returned_qty),
		gross: formatRupiah(row.split.gross),
		storeShare: formatRupiah(row.split.store_share),
		consignorShare: formatRupiah(row.split.consignor_share),
		paid: formatRupiah(row.paid)
	};
}

export interface SettlementCells extends Record<string, string> {
	key: string;
	name: string;
	issued: string;
	paid: string;
	unpaid: string;
	overdue: string;
}

type SettlementRow = SettlementReport['total'];
type Tally = SettlementRow['issued'];

function tallyText(tally: Tally): string {
	return id.reports.tally(formatInteger(tally.count), formatRupiah(tally.amount));
}

export function settlementCells(row: SettlementRow): SettlementCells {
	return {
		key: row.partner_id,
		name: row.name,
		issued: tallyText(row.issued),
		paid: tallyText(row.paid),
		unpaid: tallyText(row.unpaid),
		overdue: tallyText(row.overdue)
	};
}
