export interface ChartBox {
	width: number;
	height: number;
}

export interface AreaShape {
	line: string;
	area: string;
	ceiling: number;
}

const tenths = 10;

function round(value: number): number {
	return Math.round(value * tenths) / tenths;
}

export function areaShape(values: number[], box: ChartBox): AreaShape {
	const ceiling = Math.max(1, ...values);
	const lastIndex = Math.max(1, values.length - 1);
	const points = values.map((value, index) => ({
		left: round((index / lastIndex) * box.width),
		top: round(box.height - (value / ceiling) * box.height)
	}));
	if (points.length === 0) {
		return { line: '', area: '', ceiling };
	}
	const line = points.map(
		(point, index) => `${index === 0 ? 'M' : 'L'}${String(point.left)},${String(point.top)}`
	);
	const lastLeft = points.at(-1)?.left ?? 0;
	const floor = String(box.height);
	const area = `${line.join(' ')} L${String(lastLeft)},${floor} L0,${floor} Z`;
	return { line: line.join(' '), area, ceiling };
}
