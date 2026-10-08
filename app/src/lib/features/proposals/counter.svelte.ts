import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { counterProposal, type Proposal } from '#lib/api/proposals.js';
import type { SendOutcome } from './composer.svelte';
import {
	emptyInvitation,
	toInvitationRequest,
	validateInvitation,
	type InvitationErrors,
	type InvitationValues,
	type ItemValues
} from './schema';

const percentPerBasisPoint = 100;

type ProposalItem = Proposal['revision']['items'][number];
type ItemAction = ProposalItem['action'];

function itemOf(item: ProposalItem): ItemValues {
	return {
		productId: item.product_id,
		name: item.product_name,
		unit: item.unit,
		model: item.pricing_model,
		retailPrice: item.retail_price,
		depositPrice: item.deposit_price,
		commissionPercent: String(item.commission_bps / percentPerBasisPoint)
	};
}

export class CounterComposer {
	values = $state<InvitationValues>(emptyInvitation());
	errors = $state<InvitationErrors>({});
	busy = $state(false);
	#actions: Record<string, ItemAction> = {};
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	start(proposal: Proposal): void {
		const { terms, items } = proposal.revision;
		this.#actions = Object.fromEntries(items.map((item) => [item.product_id, item.action]));
		this.values = {
			schedule: terms.settlement_schedule,
			paymentTermsDays: terms.payment_terms_days,
			rhythm: terms.count_rhythm,
			notes: terms.notes,
			message: '',
			items: items.map(itemOf)
		};
		this.errors = {};
	}

	actionOf(productId: string): ItemAction {
		return this.#actions[productId] ?? 'ADD';
	}

	remove(productId: string): void {
		this.values.items = this.values.items.filter((item) => item.productId !== productId);
	}

	async send(proposalId: string): Promise<SendOutcome> {
		this.errors = validateInvitation(this.values);
		if (Object.keys(this.errors).length > 0) {
			return { kind: 'invalid' };
		}
		const request = toInvitationRequest(this.values);
		request.items = request.items.map((item) => ({
			...item,
			action: this.actionOf(item.product_id)
		}));
		this.busy = true;
		try {
			return { kind: 'sent', proposal: await counterProposal(this.#client(), proposalId, request) };
		} catch (error) {
			return this.#explain(error);
		} finally {
			this.busy = false;
		}
	}

	#explain(error: unknown): SendOutcome {
		if (!(error instanceof ApiError) || error.fields.length === 0) {
			return { kind: 'failed', message: error instanceof Error ? error.message : '' };
		}
		this.errors = Object.fromEntries(
			error.fields.map((problem) => [problem.field, problem.message])
		);
		return { kind: 'invalid' };
	}
}
