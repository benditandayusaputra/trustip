import { describe, expect, it } from 'vitest';
import { columnChart, donutSlices, stackedBars } from './chart';

describe('column chart', () => {
	it('FR-RPT-05 scales each day to the busiest day and centres columns in equal slots', () => {
		expect(columnChart([0, 10, 20, 40])).toEqual({
			ceiling: 40,
			columns: [
				{ left: 4.5, width: 16, top: 100, height: 0 },
				{ left: 29.5, width: 16, top: 75, height: 25 },
				{ left: 54.5, width: 16, top: 50, height: 50 },
				{ left: 79.5, width: 16, top: 0, height: 100 }
			]
		});
	});

	it('keeps every column on the floor when nothing was sold', () => {
		expect(columnChart([0, 0]).columns.map((column) => column.height)).toEqual([0, 0]);
	});

	it('draws nothing without days', () => {
		expect(columnChart([])).toEqual({ ceiling: 1, columns: [] });
	});
});

describe('donut slices', () => {
	it('splits the ring by share and starts each slice where the previous one ended', () => {
		expect(donutSlices([1500, 4500], 100)).toEqual([
			{ length: 25, offset: 0 },
			{ length: 75, offset: -25 }
		]);
	});

	it('leaves the ring empty when every value is zero', () => {
		expect(donutSlices([0, 0], 100)).toEqual([
			{ length: 0, offset: 0 },
			{ length: 0, offset: 0 }
		]);
	});
});

describe('stacked bars', () => {
	it('measures every segment against the longest row and lines segments up end to end', () => {
		expect(stackedBars([[30, 10], [20], [0, 0]])).toEqual([
			[
				{ start: 0, width: 75 },
				{ start: 75, width: 25 }
			],
			[{ start: 0, width: 50 }],
			[
				{ start: 0, width: 0 },
				{ start: 0, width: 0 }
			]
		]);
	});

	it('keeps bars empty when nothing was sold', () => {
		expect(stackedBars([[0], [0]])).toEqual([[{ start: 0, width: 0 }], [{ start: 0, width: 0 }]]);
	});
});
