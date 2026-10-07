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

export type ProfileUpdateRequest = components['schemas']['ProfileUpdateRequest'];

export function getOrganization(client: ApiClient, organizationId: string): Promise<Organization> {
	return client.get<Organization>(`/api/v1/organizations/${encodeURIComponent(organizationId)}`);
}

export function updateProfile(
	client: ApiClient,
	organizationId: string,
	request: ProfileUpdateRequest
): Promise<Organization> {
	return client.update<Organization>(
		`/api/v1/organizations/${encodeURIComponent(organizationId)}`,
		request
	);
}

export type StoreProfile = components['schemas']['StoreProfile'];
export type StoreProfileRequest = components['schemas']['StoreProfileRequest'];
export type StoreType = components['schemas']['StoreType'];
export type Category = components['schemas']['Category'];

export async function listCategories(client: ApiClient): Promise<Category[]> {
	return (await client.list<Category>('/api/v1/categories')).items;
}

export function updateStoreProfile(
	client: ApiClient,
	organizationId: string,
	request: StoreProfileRequest
): Promise<Organization> {
	return client.update<Organization>(
		`/api/v1/organizations/${encodeURIComponent(organizationId)}/store-profile`,
		request
	);
}

export type ConsignorProfile = components['schemas']['ConsignorProfile'];
export type PaymentInfoRequest = components['schemas']['PaymentInfoRequest'];

export function updateConsignorProfile(
	client: ApiClient,
	organizationId: string,
	request: ConsignorProfile
): Promise<Organization> {
	return client.update<Organization>(
		`/api/v1/organizations/${encodeURIComponent(organizationId)}/consignor-profile`,
		request
	);
}

export function updatePaymentInfo(
	client: ApiClient,
	organizationId: string,
	request: PaymentInfoRequest
): Promise<Organization> {
	return client.update<Organization>(
		`/api/v1/organizations/${encodeURIComponent(organizationId)}/payment-info`,
		request
	);
}
