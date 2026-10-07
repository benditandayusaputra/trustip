import type { ApiClient } from './client';
import type { components } from './schema';

export type Agreement = components['schemas']['Agreement'];
export type AgreementVersion = components['schemas']['AgreementVersion'];

function agreementsPath(partnershipId: string): string {
	return `/api/v1/partnerships/${encodeURIComponent(partnershipId)}`;
}

export function getAgreement(
	client: ApiClient,
	partnershipId: string,
	version: number
): Promise<Agreement> {
	const suffix = version > 0 ? `/agreements/${String(version)}` : '/agreement';
	return client.get<Agreement>(`${agreementsPath(partnershipId)}${suffix}`);
}

export async function listAgreementVersions(
	client: ApiClient,
	partnershipId: string
): Promise<AgreementVersion[]> {
	return (await client.list<AgreementVersion>(`${agreementsPath(partnershipId)}/agreements`)).items;
}
