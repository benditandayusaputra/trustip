import type { ApiClient } from './client';
import type { components } from './schema';

export type SalesReport = components['schemas']['SalesReport'];
export type ReportGrouping = SalesReport['group_by'];

export interface SalesFilter {
	from: string;
	to: string;
	groupBy: ReportGrouping;
}

export function getSalesReport(client: ApiClient, filter: SalesFilter): Promise<SalesReport> {
	const query = new URLSearchParams({ from: filter.from, to: filter.to, group_by: filter.groupBy });
	return client.get<SalesReport>(`/api/v1/reports/sales?${query.toString()}`);
}
