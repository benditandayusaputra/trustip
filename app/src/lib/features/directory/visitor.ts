import type { SearchQuery } from './query';

const coordinateDigits = 5;

export function forVisitor<T extends SearchQuery>(query: T, isVisitor: boolean): T {
	return isVisitor ? { ...query, radius: 'all' } : query;
}

export function isRandomOrder(query: SearchQuery, isVisitor: boolean): boolean {
	return (
		isVisitor &&
		query.origin === 'ORGANIZATION' &&
		(query.sort === 'NEAREST' || query.sort === 'BEST_MATCH')
	);
}

export function shuffled<T>(items: T[], random: () => number = Math.random): T[] {
	const copy = [...items];
	for (let index = copy.length - 1; index > 0; index--) {
		const other = Math.floor(random() * (index + 1));
		[copy[index], copy[other]] = [copy[other] as T, copy[index] as T];
	}
	return copy;
}

export function positionOf(point: {
	latitude: number;
	longitude: number;
}): Pick<SearchQuery, 'origin' | 'lat' | 'lng'> {
	return {
		origin: 'HERE',
		lat: point.latitude.toFixed(coordinateDigits),
		lng: point.longitude.toFixed(coordinateDigits)
	};
}
