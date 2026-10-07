import type { ApiClient } from './client';
import type { components } from './schema';

export type OnboardingRequest = components['schemas']['OnboardingRequest'];
export type Organization = components['schemas']['Organization'];
export type OrganizationType = components['schemas']['OrganizationType'];
export type Province = components['schemas']['Province'];
export type City = components['schemas']['City'];

export function createOrganization(
	client: ApiClient,
	request: OnboardingRequest
): Promise<Organization> {
	return client.send<Organization>('/api/v1/organizations', request);
}

export async function listProvinces(client: ApiClient): Promise<Province[]> {
	return (await client.list<Province>('/api/v1/provinces')).items;
}

export async function listCities(client: ApiClient, provinceCode: string): Promise<City[]> {
	const query = new URLSearchParams({ province_code: provinceCode }).toString();
	return (await client.list<City>(`/api/v1/cities?${query}`)).items;
}
