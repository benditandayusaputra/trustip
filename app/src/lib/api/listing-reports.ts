import type { ApiClient } from './client';
import type { components } from './schema';

export type ListingReportRequest = components['schemas']['ListingReportRequest'];
type ListingReport = components['schemas']['ListingReportResponse']['data'];

export function reportListing(
	client: ApiClient,
	request: ListingReportRequest
): Promise<ListingReport> {
	return client.send<ListingReport>('/api/v1/listing-reports', request);
}
