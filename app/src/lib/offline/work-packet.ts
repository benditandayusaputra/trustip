import { getAccount } from '#lib/api/account.js';
import { getAgreement } from '#lib/api/agreements.js';
import type { ApiClient } from '#lib/api/client.js';
import { getCountSheet } from '#lib/api/counts.js';
import { listPartnerCatalog, listPartnerships } from '#lib/api/partnerships.js';

const activeAgreement = 0;

export async function refreshWorkPackets(client: ApiClient): Promise<number> {
	const account = await getAccount(client);
	if (account.membership?.organization_type !== 'STORE') {
		return 0;
	}
	const partnershipIds = await listActivePartnerships(client);
	await Promise.all(partnershipIds.map((partnershipId) => fetchWorkPacket(client, partnershipId)));
	return partnershipIds.length;
}

async function listActivePartnerships(client: ApiClient): Promise<string[]> {
	const ids: string[] = [];
	let cursor: string | null = null;
	do {
		const page = await listPartnerships(client, 'ACTIVE', cursor);
		ids.push(...page.items.map((partnership) => partnership.id));
		cursor = page.hasMore ? page.nextCursor : null;
	} while (cursor !== null);
	return ids;
}

async function fetchWorkPacket(client: ApiClient, partnershipId: string): Promise<void> {
	await Promise.all([
		getAgreement(client, partnershipId, activeAgreement),
		listPartnerCatalog(client, partnershipId),
		getCountSheet(client, partnershipId)
	]);
}
