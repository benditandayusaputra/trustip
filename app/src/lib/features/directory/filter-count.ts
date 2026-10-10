import { defaultStoreQuery, type StoreQuery } from './query';

export function activeStoreFilters(query: StoreQuery): number {
	const defaults = defaultStoreQuery();
	const fields = [
		'origin',
		'radius',
		'storeType',
		'pricingModel',
		'maxCommissionBps',
		'withoutPermit',
		'withoutHalal',
		'hasOpenNeed'
	] as const;
	return (
		fields.filter((field) => query[field] !== defaults[field]).length +
		Number(query.categoryIds !== null)
	);
}
