import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import {
	claimInvitation,
	getProposal,
	viewInvitation,
	type Invitation,
	type Proposal
} from '#lib/api/proposals.js';

export type LoadStatus = 'loading' | 'ready' | 'failed';

export class ProposalRecord {
	proposal = $state<Proposal>();
	status = $state<LoadStatus>('loading');
	failure = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(proposalId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.proposal = await getProposal(this.#client(), proposalId);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}
}

export type LandingStatus = 'loading' | 'ready' | 'signedOut' | 'newcomer' | 'refused';

export class InvitationLanding {
	invitation = $state<Invitation>();
	status = $state<LandingStatus>('loading');
	message = $state('');
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(code: string): Promise<void> {
		this.status = 'loading';
		try {
			this.invitation = await viewInvitation(this.#client(), code);
			this.status = 'ready';
		} catch (error) {
			this.#refuse(error);
		}
	}

	async claim(code: string): Promise<Proposal | undefined> {
		this.busy = true;
		try {
			return await claimInvitation(this.#client(), code);
		} catch (error) {
			this.#refuse(error);
			return undefined;
		} finally {
			this.busy = false;
		}
	}

	#refuse(error: unknown): void {
		if (error instanceof ApiError && error.isSessionProblem()) {
			this.status = 'signedOut';
			return;
		}
		const isNewcomer = error instanceof ApiError && error.code === 'FORBIDDEN' && !this.invitation;
		this.status = isNewcomer ? 'newcomer' : 'refused';
		this.message = error instanceof Error ? error.message : '';
	}
}
