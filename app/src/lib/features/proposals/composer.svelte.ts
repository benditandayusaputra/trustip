import type { ApiClient } from '#lib/api/client.js';
import { listDirectory, type DirectoryProductCard } from '#lib/api/directory.js';
import { ApiError } from '#lib/api/errors.js';
import { listProducts } from '#lib/api/products.js';
import { sendInvitation, sendProposal, type Proposal } from '#lib/api/proposals.js';
import {
	emptyInvitation,
	entryOfProduct,
	itemFromEntry,
	toInvitationRequest,
	validateInvitation,
	type CatalogEntry,
	type InvitationErrors,
	type InvitationValues
} from './schema';

export type SendOutcome =
	{ kind: 'sent'; proposal: Proposal } | { kind: 'invalid' } | { kind: 'failed'; message: string };

function entryOfListing(card: DirectoryProductCard): CatalogEntry {
	return {
		id: card.id,
		name: card.name,
		unit: card.unit,
		retailPrice: card.suggested_retail_price,
		depositPrice: undefined
	};
}

export class InvitationComposer {
	values = $state<InvitationValues>(emptyInvitation());
	catalog = $state<CatalogEntry[]>([]);
	errors = $state<InvitationErrors>({});
	busy = $state(false);
	readonly #client: () => ApiClient;
	readonly #targetId: () => string;

	constructor(client: () => ApiClient, targetId: () => string = () => '') {
		this.#client = client;
		this.#targetId = targetId;
	}

	async loadCatalog(): Promise<void> {
		const search = { text: '', categoryId: '', status: 'ACTIVE' as const };
		const page = await listProducts(this.#client(), search, null).catch(() => undefined);
		this.catalog = (page?.items ?? []).map(entryOfProduct);
	}

	async loadTargetCatalog(preselectedId: string): Promise<void> {
		const page = await listDirectory<DirectoryProductCard>(this.#client(), {
			path: '/api/v1/directory/products',
			query: `consignor_id=${encodeURIComponent(this.#targetId())}&radius_km=all`,
			cursor: null
		}).catch(() => undefined);
		this.catalog = (page?.items ?? []).map(entryOfListing);
		const preselected = this.catalog.find((entry) => entry.id === preselectedId);
		if (preselected) {
			this.choose(preselected, true);
		}
	}

	isChosen(productId: string): boolean {
		return this.values.items.some((item) => item.productId === productId);
	}

	choose(entry: CatalogEntry, isChosen: boolean): void {
		const others = this.values.items.filter((item) => item.productId !== entry.id);
		this.values.items = isChosen ? [...others, itemFromEntry(entry)] : others;
	}

	async send(): Promise<SendOutcome> {
		this.errors = validateInvitation(this.values);
		if (Object.keys(this.errors).length > 0) {
			return { kind: 'invalid' };
		}
		this.busy = true;
		try {
			return { kind: 'sent', proposal: await this.#deliver() };
		} catch (error) {
			return this.#explain(error);
		} finally {
			this.busy = false;
		}
	}

	#deliver(): Promise<Proposal> {
		const request = toInvitationRequest(this.values);
		if (this.#targetId() === '') {
			return sendInvitation(this.#client(), request);
		}
		return sendProposal(this.#client(), { ...request, target_organization_id: this.#targetId() });
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
