import type { ApiClient } from './client';
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
	return client.get<Proposal>(`/api/v1/proposals/${encodeURIComponent(proposalId)}`);
}
