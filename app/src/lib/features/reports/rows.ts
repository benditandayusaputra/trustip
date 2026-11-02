import type { DailyRecap, SalesReport } from '#lib/api/reports.js';
import type { Sheet, SheetCell } from '#lib/domain/xlsx.js';
import {
	formatBasisPoints,
	formatBusinessDate,
	formatInteger,
	formatRupiah
} from '#lib/i18n/format.js';
import { reports as id } from '#lib/i18n/reports.js';

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

const basisPointsPerWhole = 10_000;
const labelWidth = 34;
const figureWidth = 20;
const bodyStyles = {
	text: 'text',
	integer: 'integer',
	rupiah: 'rupiah',
	percent: 'percent'
} as const;
const totalStyles = {
	text: 'totalText',
	integer: 'totalInteger',
	rupiah: 'totalRupiah',
	percent: 'totalPercent'
} as const;

function sheetRow(
	label: string,
	figures: Figures,
	styles: typeof bodyStyles | typeof totalStyles
): SheetCell[] {
	return [
		{ value: label, style: styles.text },
		{ value: figures.sold_qty, style: styles.integer },
		{ value: figures.split.gross, style: styles.rupiah },
		{ value: figures.split.store_share, style: styles.rupiah },
		{ value: figures.split.consignor_share, style: styles.rupiah },
		{ value: figures.returned_qty, style: styles.integer },
		{ value: figures.returned_bps / basisPointsPerWhole, style: styles.percent }
	];
}

export function salesSheet(report: SalesReport, isStore: boolean): Sheet {
	const text = id.reports;
	const grouping = text.groupings[isStore ? 'STORE' : 'CONSIGNOR'][report.group_by];
	const period = text.sheetPeriod(formatBusinessDate(report.from), formatBusinessDate(report.to));
	const header = [
		text.rowLabel,
		text.soldUnits,
		text.gross,
		text.storeShare,
		isStore ? text.mustPay : text.consignorShare,
		text.returnedUnits,
		text.returnedPercent
	];
	const intro: SheetCell[][] = [
		[{ value: text.salesTitle, style: 'title' }],
		[{ value: `${period} · ${grouping}`, style: 'note' }],
		[],
		header.map((value) => ({ value, style: 'header' }))
	];
	return {
		name: text.tabs.sales,
		columnWidths: header.map((_, index) => (index === 0 ? labelWidth : figureWidth)),
		frozenRows: intro.length,
		rows: [
			...intro,
			...report.rows.map((row) => sheetRow(row.label, row, bodyStyles)),
			sheetRow(text.total, report.summary, totalStyles)
		]
	};
}

export interface DayCells extends Record<string, string> {
	key: string;
	day: string;
	sold: string;
	gross: string;
}

export function dayCells(report: SalesReport): DayCells[] {
	return report.rows.map((row) => ({
		key: row.key,
		day: formatBusinessDate(row.key),
		sold: formatInteger(row.sold_qty),
		gross: formatRupiah(row.split.gross)
	}));
}

export const settlementKinds = [
	{ key: 'issued', label: id.reports.issued, tone: 'text-accent', dot: 'bg-accent' },
	{ key: 'paid', label: id.reports.settled, tone: 'text-success', dot: 'bg-success' },
	{ key: 'unpaid', label: id.reports.unpaid, tone: 'text-warning', dot: 'bg-warning-fill' },
	{ key: 'overdue', label: id.reports.overdue, tone: 'text-danger', dot: 'bg-danger-fill' }
] as const;
