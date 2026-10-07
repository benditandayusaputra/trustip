import type { ApiClient } from '#lib/api/client.js';
import { listProposals, type ProposalBox, type ProposalSummary } from '#lib/api/proposals.js';
import type { LoadStatus } from './record.svelte';

export class ProposalInbox {
	box = $state<ProposalBox>('INCOMING');
	items = $state<ProposalSummary[]>([]);
	status = $state<LoadStatus>('loading');
	failure = $state('');
	nextCursor = $state<string | null>(null);
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(box: ProposalBox): Promise<void> {
		this.box = box;
		this.status = 'loading';
		try {
			const page = await listProposals(this.#client(), box, null);
			this.items = page.items;
			this.nextCursor = page.hasMore ? page.nextCursor : null;
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	async loadMore(): Promise<void> {
		if (this.nextCursor === null) {
			return;
		}
		this.busy = true;
		try {
			const page = await listProposals(this.#client(), this.box, this.nextCursor);
			this.items = [...this.items, ...page.items];
			this.nextCursor = page.hasMore ? page.nextCursor : null;
		} finally {
			this.busy = false;
		}
	}
}
