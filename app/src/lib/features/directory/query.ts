type OriginKind = 'ORGANIZATION' | 'HERE' | 'CITY';
export type RadiusChoice = '2' | '5' | '10' | '25' | 'all';
export type SortChoice = 'NEAREST' | 'BEST_MATCH' | 'REPUTATION' | 'NEWEST';

export interface StoreQuery {
	origin: OriginKind;
	lat: string;
	lng: string;
	cityCode: string;
	radius: RadiusChoice;
	sort: SortChoice;
	storeType: string;
	categoryIds: string[] | null;
	pricingModel: string;
	maxCommissionBps: string;
	withoutPermit: boolean;
	withoutHalal: boolean;
	text: string;
}

export const anyCategory = 'any';
export const radiusChoices: RadiusChoice[] = ['2', '5', '10', '25', 'all'];
export const sortChoices: SortChoice[] = ['NEAREST', 'BEST_MATCH', 'REPUTATION', 'NEWEST'];
const originChoices: OriginKind[] = ['ORGANIZATION', 'HERE', 'CITY'];

export function defaultStoreQuery(): StoreQuery {
	return {
		origin: 'ORGANIZATION',
		lat: '',
		lng: '',
		cityCode: '',
		radius: '10',
		sort: 'NEAREST',
		storeType: '',
		categoryIds: null,
		pricingModel: '',
		maxCommissionBps: '',
		withoutPermit: false,
		withoutHalal: false,
		text: ''
	};
}

function choiceOf<T extends string>(raw: string | null, choices: T[], fallback: T): T {
	return choices.find((choice) => choice === raw) ?? fallback;
}

interface QueryReader {
	get(name: string): string | null;
	getAll(name: string): string[];
}

export function storeQueryFrom(params: QueryReader): StoreQuery {
	const fallback = defaultStoreQuery();
	const categories = params.getAll('category_id');
	const textOf = (name: string) => params.get(name) ?? '';
	return {
		origin: choiceOf(params.get('origin'), originChoices, fallback.origin),
		lat: textOf('lat'),
		lng: textOf('lng'),
		cityCode: textOf('city_code'),
		radius: choiceOf(params.get('radius_km'), radiusChoices, fallback.radius),
		sort: choiceOf(params.get('sort'), sortChoices, fallback.sort),
		storeType: textOf('store_type'),
		categoryIds: categories.length === 0 ? null : categories,
		pricingModel: textOf('pricing_model'),
		maxCommissionBps: textOf('max_commission_bps'),
		withoutPermit: params.get('without_permit') === 'true',
		withoutHalal: params.get('without_halal') === 'true',
		text: textOf('q')
	};
}

function originParams(query: StoreQuery): [string, string][] {
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

export function storeSearchParams(query: StoreQuery): URLSearchParams {
	const fallback = defaultStoreQuery();
	const params = new URLSearchParams(originParams(query));
	const optional: [string, string, string][] = [
		['radius_km', query.radius, fallback.radius],
		['sort', query.sort, fallback.sort],
		['store_type', query.storeType, ''],
		['pricing_model', query.pricingModel, ''],
		['max_commission_bps', query.maxCommissionBps, ''],
		['without_permit', String(query.withoutPermit), 'false'],
		['without_halal', String(query.withoutHalal), 'false'],
		['q', query.text, '']
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
