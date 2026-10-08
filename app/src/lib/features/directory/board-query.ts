import {
	defaultSearchQuery,
	searchParamsOf,
	searchQueryFrom,
	type QueryReader,
	type SearchQuery
} from './query';

type BoardSort = 'NEAREST' | 'NEWEST';

export interface BoardQuery extends Omit<SearchQuery, 'sort'> {
	sort: BoardSort;
}

export function defaultBoardQuery(): BoardQuery {
	return { ...defaultSearchQuery(), sort: 'NEAREST' };
}

export function boardQueryFrom(params: QueryReader): BoardQuery {
	const query = searchQueryFrom(params);
	return { ...query, sort: query.sort === 'NEWEST' ? 'NEWEST' : 'NEAREST' };
}

export function boardSearchParams(query: BoardQuery): URLSearchParams {
	return searchParamsOf(query, []);
}
