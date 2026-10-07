import type { ApiClient } from './client';
import type { components } from './schema';

export type Agreement = components['schemas']['Agreement'];

export function getAgreement(client: ApiClient, partnershipId: string): Promise<Agreement> {
	return client.get<Agreement>(
		`/api/v1/partnerships/${encodeURIComponent(partnershipId)}/agreement`
	);
}
