import type { ApiClient } from '#lib/api/client.js';
import { listRevisions, type ProposalRevisionRecord } from '#lib/api/proposals.js';

type ChangedItem = ProposalRevisionRecord['items'][number];

export class RevisionTrail {
	records = $state<ProposalRevisionRecord[]>([]);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(proposalId: string): Promise<void> {
		this.records = await listRevisions(this.#client(), proposalId).catch(() => []);
	}

	get latest(): ProposalRevisionRecord | undefined {
		const last = this.records.at(-1);
		return last && last.revision_no > 1 ? last : undefined;
	}

	changeOf(productId: string): ChangedItem['change'] | undefined {
		const change = this.latest?.items.find((item) => item.product_id === productId)?.change;
		return change === 'KEPT' ? undefined : change;
	}

	get removed(): ChangedItem[] {
		return this.latest?.items.filter((item) => item.change === 'REMOVED') ?? [];
	}
}
