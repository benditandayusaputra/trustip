import { getAgreement, type Agreement } from '#lib/api/agreements.js';
import type { ApiClient } from '#lib/api/client.js';
import { getCountSheet, type CountSheet } from '#lib/api/counts.js';
import { id } from '#lib/i18n/id.js';
import type { OutboxEntry } from '#lib/offline/outbox.js';
import type { FlushResult } from '#lib/offline/sync.js';
import {
	countReviewRows,
	dropReviewRows,
	type CountReviewRow,
	type DropReviewRow
} from './review-rows';

const activeAgreement = 0;

interface ReviewDeps {
	client: () => ApiClient;
	find: (entryId: string) => Promise<OutboxEntry | undefined>;
	discard: (entryId: string) => Promise<FlushResult>;
}

export type Latest = { kind: 'COUNT'; sheet: CountSheet } | { kind: 'DROP'; agreement: Agreement };

export class EntryReview {
	entry = $state<OutboxEntry>();
	latest = $state<Latest>();
	status = $state<'loading' | 'ready' | 'missing' | 'failed'>('loading');
	failure = $state('');
	readonly #deps: ReviewDeps;

	constructor(deps: ReviewDeps) {
		this.#deps = deps;
	}

	get countRows(): CountReviewRow[] {
		const { entry, latest } = this;
		return entry?.kind === 'COUNT' && latest?.kind === 'COUNT'
			? countReviewRows(entry, latest.sheet)
			: [];
	}

	get dropRows(): DropReviewRow[] {
		const { entry, latest } = this;
		return entry?.kind === 'DROP' && latest?.kind === 'DROP'
			? dropReviewRows(entry, latest.agreement)
			: [];
	}

	get versionNote(): string {
		const { entry, latest } = this;
		return entry?.kind === 'DROP' && latest?.kind === 'DROP'
			? id.outbox.versions(entry.request.agreement_version, latest.agreement.version)
			: '';
	}

	async load(entryId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.entry = await this.#deps.find(entryId);
			if (!this.entry) {
				this.status = 'missing';
				return;
			}
			this.latest = await this.#latestFor(this.entry);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	async discard(): Promise<void> {
		if (this.entry) {
			await this.#deps.discard(this.entry.id);
		}
	}

	async #latestFor(entry: OutboxEntry): Promise<Latest> {
		const client = this.#deps.client();
		if (entry.kind === 'COUNT') {
			return { kind: 'COUNT', sheet: await getCountSheet(client, entry.partnershipId) };
		}
		return {
			kind: 'DROP',
			agreement: await getAgreement(client, entry.partnershipId, activeAgreement)
		};
	}
}
