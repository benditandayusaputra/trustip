type OriginKind = 'ORGANIZATION' | 'HERE' | 'CITY';
export type RadiusChoice = '2' | '5' | '10' | '25' | 'all';
export type SortChoice = 'NEAREST' | 'BEST_MATCH' | 'REPUTATION' | 'NEWEST';

export interface SearchQuery {
	origin: OriginKind;
	lat: string;
	lng: string;
	cityCode: string;
	radius: RadiusChoice;
	sort: SortChoice;
	categoryIds: string[] | null;
	text: string;
}

export interface StoreQuery extends SearchQuery {
	storeType: string;
	pricingModel: string;
	maxCommissionBps: string;
	withoutPermit: boolean;
	withoutHalal: boolean;
}

export interface QueryReader {
	get(name: string): string | null;
	getAll(name: string): string[];
}

export type Optional = [name: string, value: string, unset: string];

export const anyCategory = 'any';
export const radiusChoices: RadiusChoice[] = ['2', '5', '10', '25', 'all'];
export const sortChoices: SortChoice[] = ['NEAREST', 'BEST_MATCH', 'REPUTATION', 'NEWEST'];
const originChoices: OriginKind[] = ['ORGANIZATION', 'HERE', 'CITY'];

export function defaultSearchQuery(): SearchQuery {
	return {
		origin: 'ORGANIZATION',
		lat: '',
		lng: '',
		cityCode: '',
		radius: '10',
		sort: 'NEAREST',
		categoryIds: null,
		text: ''
	};
}

export function defaultStoreQuery(): StoreQuery {
	return {
		...defaultSearchQuery(),
		storeType: '',
		pricingModel: '',
		maxCommissionBps: '',
		withoutPermit: false,
		withoutHalal: false
	};
}

function choiceOf<T extends string>(raw: string | null, choices: T[], fallback: T): T {
	return choices.find((choice) => choice === raw) ?? fallback;
}

export function textOf(params: QueryReader, name: string): string {
	return params.get(name) ?? '';
}

export function searchQueryFrom(params: QueryReader): SearchQuery {
	const fallback = defaultSearchQuery();
	const categories = params.getAll('category_id');
	return {
		origin: choiceOf(params.get('origin'), originChoices, fallback.origin),
		lat: textOf(params, 'lat'),
		lng: textOf(params, 'lng'),
		cityCode: textOf(params, 'city_code'),
		radius: choiceOf(params.get('radius_km'), radiusChoices, fallback.radius),
		sort: choiceOf(params.get('sort'), sortChoices, fallback.sort),
		categoryIds: categories.length === 0 ? null : categories,
		text: textOf(params, 'q')
	};
}

export function storeQueryFrom(params: QueryReader): StoreQuery {
	return {
		...searchQueryFrom(params),
		storeType: textOf(params, 'store_type'),
		pricingModel: textOf(params, 'pricing_model'),
		maxCommissionBps: textOf(params, 'max_commission_bps'),
		withoutPermit: params.get('without_permit') === 'true',
		withoutHalal: params.get('without_halal') === 'true'
	};
}

function originParams(query: SearchQuery): [string, string][] {
	if (query.origin === 'HERE') {
		return [
			['origin', 'HERE'],
			['lat', query.lat],
			['lng', query.lng]
		];
	}
	if (query.origin === 'CITY') {
		return [
			['origin', 'CITY'],
			['city_code', query.cityCode]
		];
	}
	return [];
}

export function searchParamsOf(query: SearchQuery, extra: Optional[]): URLSearchParams {
	const fallback = defaultSearchQuery();
	const params = new URLSearchParams(originParams(query));
	const optional: Optional[] = [
		['radius_km', query.radius, fallback.radius],
		['sort', query.sort, fallback.sort],
		['q', query.text, ''],
		...extra
	];
	for (const [name, value, unset] of optional) {
		if (value !== unset) {
			params.set(name, value);
		}
	}
	for (const categoryId of query.categoryIds ?? []) {
		params.append('category_id', categoryId);
	}
	return params;
}

export function storeSearchParams(query: StoreQuery): URLSearchParams {
	return searchParamsOf(query, [
		['store_type', query.storeType, ''],
		['pricing_model', query.pricingModel, ''],
		['max_commission_bps', query.maxCommissionBps, ''],
		['without_permit', String(query.withoutPermit), 'false'],
		['without_halal', String(query.withoutHalal), 'false']
	]);
}
