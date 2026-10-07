import type { ApiClient } from './client';
import type { Page } from './envelope';
import type { components } from './schema';

export type Proposal = components['schemas']['Proposal'];
export type InvitationRequest = components['schemas']['InvitationRequest'];
export type Invitation = components['schemas']['Invitation'];
export type SettlementSchedule = components['schemas']['SettlementSchedule'];
export type CountRhythm = components['schemas']['CountRhythm'];

const invitationsPath = '/api/v1/invitations';

function invitationPath(code: string): string {
	return `${invitationsPath}/${encodeURIComponent(code)}`;
}

export function sendInvitation(client: ApiClient, request: InvitationRequest): Promise<Proposal> {
	return client.send<Proposal>(invitationsPath, request);
}

export function viewInvitation(client: ApiClient, code: string): Promise<Invitation> {
	return client.get<Invitation>(invitationPath(code));
}

export function claimInvitation(client: ApiClient, code: string): Promise<Proposal> {
	return client.send<Proposal>(`${invitationPath(code)}/claim`);
}

export function getProposal(client: ApiClient, proposalId: string): Promise<Proposal> {
	return client.get<Proposal>(proposalPath(proposalId));
}

export type ProposalSummary = components['schemas']['ProposalSummary'];
export type ProposalBox = 'INCOMING' | 'OUTGOING';

function proposalPath(proposalId: string): string {
	return `/api/v1/proposals/${encodeURIComponent(proposalId)}`;
}

export function acceptProposal(client: ApiClient, proposalId: string): Promise<Proposal> {
	return client.send<Proposal>(`${proposalPath(proposalId)}/accept`);
}

export function declineProposal(
	client: ApiClient,
	proposalId: string,
	reason: string
): Promise<Proposal> {
	return client.send<Proposal>(`${proposalPath(proposalId)}/decline`, { reason });
}

export function withdrawProposal(client: ApiClient, proposalId: string): Promise<Proposal> {
	return client.send<Proposal>(`${proposalPath(proposalId)}/withdraw`);
}

export function listProposals(
	client: ApiClient,
	box: ProposalBox,
	cursor: string | null
): Promise<Page<ProposalSummary>> {
	const query = new URLSearchParams({ box });
	if (cursor) {
		query.set('cursor', cursor);
	}
	return client.list<ProposalSummary>(`/api/v1/proposals?${query.toString()}`);
}
