import type { ApiClient } from './client';
import type { Page } from './envelope';
import type { components } from './schema';

export type Partnership = components['schemas']['Partnership'];
export type PartnershipSummary = components['schemas']['PartnershipSummary'];
export type PartnerCatalogItem = components['schemas']['PartnerCatalogItem'];
export type PartnershipBlocker = components['schemas']['PartnershipBlocker'];

const partnershipsPath = '/api/v1/partnerships';

function partnershipPath(partnershipId: string): string {
	return `${partnershipsPath}/${encodeURIComponent(partnershipId)}`;
}

export function listPartnerships(
	client: ApiClient,
	status: PartnershipSummary['status'] | '',
	cursor: string | null
): Promise<Page<PartnershipSummary>> {
	const query = new URLSearchParams();
	if (status !== '') {
		query.set('status', status);
	}
	if (cursor) {
		query.set('cursor', cursor);
	}
	const suffix = query.size > 0 ? `?${query.toString()}` : '';
	return client.list<PartnershipSummary>(`${partnershipsPath}${suffix}`);
}

export function getPartnership(client: ApiClient, partnershipId: string): Promise<Partnership> {
	return client.get<Partnership>(partnershipPath(partnershipId));
}

export async function listPartnerCatalog(
	client: ApiClient,
	partnershipId: string
): Promise<PartnerCatalogItem[]> {
	return (await client.list<PartnerCatalogItem>(`${partnershipPath(partnershipId)}/catalog`)).items;
}

export async function listBlockers(
	client: ApiClient,
	partnershipId: string
): Promise<PartnershipBlocker[]> {
	return (await client.list<PartnershipBlocker>(`${partnershipPath(partnershipId)}/blockers`))
		.items;
}

export function endPartnership(
	client: ApiClient,
	partnershipId: string,
	reason: string
): Promise<Partnership> {
	return client.send<Partnership>(`${partnershipPath(partnershipId)}/end`, { reason });
}

export function setPublicListing(
	client: ApiClient,
	partnershipId: string,
	showOnPublic: boolean
): Promise<Partnership> {
	return client.update<Partnership>(`${partnershipPath(partnershipId)}/public-listing`, {
		show_on_public: showOnPublic
	});
}
