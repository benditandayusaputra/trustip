import type { ApiClient } from './client';
import type { components } from './schema';

export type SalesReport = components['schemas']['SalesReport'];
export type ReportGrouping = SalesReport['group_by'];

export interface SalesNarrowing {
	partnerId?: string;
	productId?: string;
}

export interface SalesFilter extends SalesNarrowing {
	from: string;
	to: string;
	groupBy: ReportGrouping;
}

export function getSalesReport(client: ApiClient, filter: SalesFilter): Promise<SalesReport> {
	const query = new URLSearchParams({ from: filter.from, to: filter.to, group_by: filter.groupBy });
	if (filter.partnerId) {
		query.set('partner_id', filter.partnerId);
	}
	if (filter.productId) {
		query.set('product_id', filter.productId);
	}
	return client.get<SalesReport>(`/api/v1/reports/sales?${query.toString()}`);
}

export type DailyRecap = components['schemas']['DailyRecapResponse']['data'];
export type SettlementReport = components['schemas']['SettlementReportResponse']['data'];

export function getDailyRecap(client: ApiClient, date: string): Promise<DailyRecap> {
	return client.get<DailyRecap>(
		`/api/v1/reports/daily?${new URLSearchParams({ date }).toString()}`
	);
}

export function getSettlementReport(
	client: ApiClient,
	range: { from: string; to: string }
): Promise<SettlementReport> {
	return client.get<SettlementReport>(
		`/api/v1/reports/settlements?${new URLSearchParams(range).toString()}`
	);
}
