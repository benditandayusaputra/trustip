import type { Proposal } from '#lib/api/proposals.js';

export function invitationUrl(origin: string, inviteCode: string): string {
	return `${origin}/undangan/${encodeURIComponent(inviteCode)}`;
}

export function whatsAppShareUrl(message: string): string {
	return `https://wa.me/?${new URLSearchParams({ text: message }).toString()}`;
}

export function senderName(proposal: Proposal): string {
	const parties = [proposal.consignor, proposal.store];
	return parties.find((party) => party?.id === proposal.from_org_id)?.name ?? '';
}

export function partyName(proposal: Proposal, organizationId: string): string {
	const parties = [proposal.consignor, proposal.store];
	return parties.find((party) => party?.id === organizationId)?.name ?? '';
}
