export interface PriceRange {
	key: string;
	minPrice: string;
	maxPrice: string;
}

export const priceRanges: PriceRange[] = [
	{ key: 'any', minPrice: '', maxPrice: '' },
	{ key: 'under5k', minPrice: '', maxPrice: '4999' },
	{ key: 'from5kTo15k', minPrice: '5000', maxPrice: '15000' },
	{ key: 'from15kTo50k', minPrice: '15000', maxPrice: '50000' },
	{ key: 'over50k', minPrice: '50001', maxPrice: '' }
];

export function priceRangeOf(minPrice: string, maxPrice: string): string {
	const range = priceRanges.find(
		(candidate) => candidate.minPrice === minPrice && candidate.maxPrice === maxPrice
	);
	return range?.key ?? 'any';
}
