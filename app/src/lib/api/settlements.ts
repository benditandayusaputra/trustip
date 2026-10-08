import type { ApiClient } from './client';
import type { Page } from './envelope';
import type { components } from './schema';

export type Settlement = components['schemas']['Settlement'];
export type SettlementSummary = components['schemas']['SettlementSummary'];
type SettlementAge = components['schemas']['SettlementAge'];
export type RunningSales = components['schemas']['RunningSales'];

export interface SettlementFilter {
	partnershipId: string;
	status: '' | 'ISSUED' | 'PAID';
	age: '' | SettlementAge;
}

const settlementsPath = '/api/v1/settlements';

export function getSettlement(client: ApiClient, settlementId: string): Promise<Settlement> {
	return client.get<Settlement>(`${settlementsPath}/${encodeURIComponent(settlementId)}`);
}

export function getRunningSales(client: ApiClient, partnershipId: string): Promise<RunningSales> {
	return client.get<RunningSales>(
		`/api/v1/partnerships/${encodeURIComponent(partnershipId)}/running-sales`
	);
}

export function listSettlements(
	client: ApiClient,
	filter: SettlementFilter,
	cursor: string | null
): Promise<Page<SettlementSummary>> {
	const query = new URLSearchParams();
	const entries: [string, string][] = [
		['partnership_id', filter.partnershipId],
		['status', filter.status],
		['age', filter.age],
		['cursor', cursor ?? '']
	];
	for (const [name, value] of entries) {
		if (value !== '') {
			query.set(name, value);
		}
	}
	const suffix = query.size > 0 ? `?${query.toString()}` : '';
	return client.list<SettlementSummary>(`${settlementsPath}${suffix}`);
}

export type Payment = components['schemas']['Payment'];
export type PaymentRequest = components['schemas']['PaymentRequest'];
export type PaymentMethod = components['schemas']['PaymentMethod'];

export interface PaymentSubmission {
	settlementId: string;
	request: PaymentRequest;
	idempotencyKey: string;
}

export function recordPayment(
	client: ApiClient,
	submission: PaymentSubmission
): Promise<Settlement> {
	return client.create<Settlement>(
		`${settlementsPath}/${encodeURIComponent(submission.settlementId)}/payments`,
		submission.request,
		submission.idempotencyKey
	);
}

export function markSettlementPaid(client: ApiClient, settlementId: string): Promise<Settlement> {
	return client.send<Settlement>(
		`${settlementsPath}/${encodeURIComponent(settlementId)}/mark-paid`
	);
}

export function confirmPayment(client: ApiClient, paymentId: string): Promise<Settlement> {
	return client.send<Settlement>(`/api/v1/payments/${encodeURIComponent(paymentId)}/confirm`);
}

export function rejectPayment(
	client: ApiClient,
	paymentId: string,
	reason: string
): Promise<Settlement> {
	return client.send<Settlement>(`/api/v1/payments/${encodeURIComponent(paymentId)}/reject`, {
		reason
	});
}
