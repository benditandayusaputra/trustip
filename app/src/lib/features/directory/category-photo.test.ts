import { describe, expect, it } from 'vitest';
import { categoryPhoto } from './category-photo';

describe('category photo', () => {
	it('points to the photo named after the category slug', () => {
		expect(categoryPhoto('Oleh-oleh khas daerah')).toBe('/kategori/oleh-oleh-khas-daerah.jpg');
		expect(categoryPhoto('Minuman')).toBe('/kategori/minuman.jpg');
	});
});
