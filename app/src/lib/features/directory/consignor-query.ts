import {
	defaultSearchQuery,
	searchParamsOf,
	searchQueryFrom,
	textOf,
	type QueryReader,
	type SearchQuery
} from './query';

export type ConsignorTab = 'produk' | 'penitip';

export interface ConsignorQuery extends SearchQuery {
	tab: ConsignorTab;
	minPrice: string;
	maxPrice: string;
	minShelfLifeDays: string;
	hasPermit: boolean;
	hasHalal: boolean;
	canSample: boolean;
	deliversToMe: boolean;
}

export function defaultConsignorQuery(): ConsignorQuery {
	return {
		...defaultSearchQuery(),
		tab: 'produk',
		minPrice: '',
		maxPrice: '',
		minShelfLifeDays: '',
		hasPermit: false,
		hasHalal: false,
		canSample: false,
		deliversToMe: false
	};
}

export function consignorQueryFrom(params: QueryReader): ConsignorQuery {
	const isTrue = (name: string) => params.get(name) === 'true';
	return {
		...searchQueryFrom(params),
		tab: params.get('tab') === 'penitip' ? 'penitip' : 'produk',
		minPrice: textOf(params, 'min_price'),
		maxPrice: textOf(params, 'max_price'),
		minShelfLifeDays: textOf(params, 'min_shelf_life_days'),
		hasPermit: isTrue('has_permit'),
		hasHalal: isTrue('has_halal'),
		canSample: isTrue('can_sample'),
		deliversToMe: isTrue('delivers_to_me')
	};
}

export function consignorSearchParams(query: ConsignorQuery): URLSearchParams {
	return searchParamsOf(query, [
		['min_price', query.minPrice, ''],
		['max_price', query.maxPrice, ''],
		['min_shelf_life_days', query.minShelfLifeDays, ''],
		['has_permit', String(query.hasPermit), 'false'],
		['has_halal', String(query.hasHalal), 'false'],
		['can_sample', String(query.canSample), 'false'],
		['delivers_to_me', String(query.deliversToMe), 'false']
	]);
}

export function consignorAddress(query: ConsignorQuery): URLSearchParams {
	const params = consignorSearchParams(query);
	if (query.tab !== 'produk') {
		params.set('tab', query.tab);
	}
	return params;
}
