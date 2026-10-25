import type { SalesReport } from '#lib/api/reports.js';
import { formatBasisPoints, formatInteger, formatRupiah } from '#lib/i18n/format.js';
import { reports as id } from '#lib/i18n/reports.js';
import type { BarRow, LegendEntry } from './components/BarChart.svelte';
import type { ChartSlice } from './components/DonutChart.svelte';

type Figures = SalesReport['summary'];

const text = id.reports;
const basisPointsPerWhole = 10_000;
const chartRowLimit = 6;

function basisPointsOf(part: number, whole: number): number {
	return whole === 0 ? 0 : Math.round((part * basisPointsPerWhole) / whole);
}

export function splitSlices(figures: Figures, isStore: boolean): ChartSlice[] {
	const { gross, store_share: storeShare, consignor_share: consignorShare } = figures.split;
	const shareLabel = (amount: number) =>
		text.shareOfGross(formatBasisPoints(basisPointsOf(amount, gross)));
	return [
		{
			key: 'consignor',
			label: isStore ? text.mustPay : text.consignorShare,
			value: consignorShare,
			valueLabel: formatRupiah(consignorShare),
			shareLabel: shareLabel(consignorShare),
			tone: 'primary'
		},
		{
			key: 'store',
			label: text.storeShare,
			value: storeShare,
			valueLabel: formatRupiah(storeShare),
			shareLabel: shareLabel(storeShare),
			tone: 'success'
		}
	];
}

export function grossBars(report: SalesReport): BarRow[] {
	return report.rows
		.filter((row) => row.sold_qty > 0)
		.slice(0, chartRowLimit)
		.map((row) => {
			const gross = formatRupiah(row.split.gross);
			return {
				key: row.key,
				label: row.label,
				valueLabel: gross,
				segments: [
					{ label: text.gross, value: row.split.gross, valueLabel: gross, tone: 'primary' }
				]
			};
		});
}

export const returnLegend: LegendEntry[] = [
	{ label: text.soldUnits, tone: 'primary' },
	{ label: text.returnedUnits, tone: 'warning' }
];

export function returnBars(report: SalesReport): BarRow[] {
	return report.rows
		.filter((row) => row.sold_qty + row.returned_qty > 0)
		.slice(0, chartRowLimit)
		.map((row) => {
			const sold = formatInteger(row.sold_qty);
			const returned = formatInteger(row.returned_qty);
			return {
				key: row.key,
				label: row.label,
				valueLabel: text.soldReturned(sold, returned),
				segments: [
					{ label: text.soldUnits, value: row.sold_qty, valueLabel: sold, tone: 'primary' },
					{
						label: text.returnedUnits,
						value: row.returned_qty,
						valueLabel: returned,
						tone: 'warning'
					}
				]
			};
		});
}
