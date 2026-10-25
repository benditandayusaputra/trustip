const tenths = 10;
const wholePercent = 100;
const columnShare = 0.64;
const half = 0.5;

function round(value: number): number {
	return Math.round(value * tenths) / tenths;
}

interface ChartColumn {
	left: number;
	width: number;
	top: number;
	height: number;
}

export interface ColumnChartShape {
	ceiling: number;
	columns: ChartColumn[];
}

export function columnChart(values: number[]): ColumnChartShape {
	const ceiling = Math.max(1, ...values);
	const slot = wholePercent / Math.max(1, values.length);
	const width = round(slot * columnShare);
	const columns = values.map((value, index) => {
		const height = round((value / ceiling) * wholePercent);
		const left = round(index * slot + (slot - width) * half);
		return { left, width, top: round(wholePercent - height), height };
	});
	return { ceiling, columns };
}

export interface DonutSlice {
	length: number;
	offset: number;
}

function sumOf(values: number[]): number {
	return values.reduce((sum, value) => sum + value, 0);
}

export function donutSlices(values: number[], circumference: number): DonutSlice[] {
	const total = sumOf(values);
	let start = 0;
	return values.map((value) => {
		const length = total === 0 ? 0 : round((value / total) * circumference);
		const slice = { length, offset: start === 0 ? 0 : -start };
		start += length;
		return slice;
	});
}

export interface BarPiece {
	start: number;
	width: number;
}

function piecesOf(widths: number[]): BarPiece[] {
	let start = 0;
	return widths.map((width) => {
		const piece = { start, width };
		start = round(start + width);
		return piece;
	});
}

export function stackedBars(rows: number[][]): BarPiece[][] {
	const ceiling = Math.max(1, ...rows.map(sumOf));
	return rows.map((row) => piecesOf(row.map((value) => round((value / ceiling) * wholePercent))));
}
