import { describe, expect, it } from 'vitest';
import { areaShape } from './chart';

const box = { width: 100, height: 50 };

describe('area chart shape', () => {
	it('FR-RPT-05 scales values to the highest day and closes the area at the bottom', () => {
		const shape = areaShape([0, 20, 40], box);

		expect(shape.ceiling).toBe(40);
		expect(shape.line).toBe('M0,50 L50,25 L100,0');
		expect(shape.area).toBe('M0,50 L50,25 L100,0 L100,50 L0,50 Z');
	});

	it('keeps a flat line on the floor when nothing was sold', () => {
		expect(areaShape([0, 0], box).line).toBe('M0,50 L100,50');
	});

	it('draws nothing without days', () => {
		expect(areaShape([], box)).toEqual({ line: '', area: '', ceiling: 1 });
	});
});
