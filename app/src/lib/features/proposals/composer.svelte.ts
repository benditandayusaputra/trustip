import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { listProducts, type Product } from '#lib/api/products.js';
import { sendInvitation, type Proposal } from '#lib/api/proposals.js';
import {
	emptyInvitation,
	itemFromProduct,
	toInvitationRequest,
	validateInvitation,
	type InvitationErrors,
	type InvitationValues
} from './schema';

export type SendOutcome =
	{ kind: 'sent'; proposal: Proposal } | { kind: 'invalid' } | { kind: 'failed'; message: string };

export class InvitationComposer {
	values = $state<InvitationValues>(emptyInvitation());
	catalog = $state<Product[]>([]);
	errors = $state<InvitationErrors>({});
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async loadCatalog(): Promise<void> {
		const search = { text: '', categoryId: '', status: 'ACTIVE' as const };
		const page = await listProducts(this.#client(), search, null).catch(() => undefined);
		this.catalog = page?.items ?? [];
	}

	isChosen(productId: string): boolean {
		return this.values.items.some((item) => item.productId === productId);
	}

	choose(product: Product, isChosen: boolean): void {
		const others = this.values.items.filter((item) => item.productId !== product.id);
		this.values.items = isChosen ? [...others, itemFromProduct(product)] : others;
	}

	async send(): Promise<SendOutcome> {
		this.errors = validateInvitation(this.values);
		if (Object.keys(this.errors).length > 0) {
			return { kind: 'invalid' };
		}
		this.busy = true;
		try {
			const proposal = await sendInvitation(this.#client(), toInvitationRequest(this.values));
			return { kind: 'sent', proposal };
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
