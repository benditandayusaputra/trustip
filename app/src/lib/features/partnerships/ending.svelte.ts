import type { ApiClient } from '#lib/api/client.js';
import {
	endPartnership,
	listBlockers,
	setPublicListing,
	type Partnership,
	type PartnershipBlocker
} from '#lib/api/partnerships.js';

type Outcome = { kind: 'done'; partnership: Partnership } | { kind: 'failed'; message: string };

const documentLinks: Record<PartnershipBlocker['kind'], string> = {
	STOCK_AT_STORE: '',
	PENDING_DROP: '/app/titipan/',
	PENDING_COUNT: '/app/hitung/',
	PENDING_PAYMENT: '/app/setoran/',
	UNPAID_SETTLEMENT: '/app/setoran/'
};

export function blockerHref(blocker: PartnershipBlocker): string | undefined {
	const base = documentLinks[blocker.kind];
	return base === '' ? undefined : `${base}${blocker.document_id}`;
}

function failed(error: unknown): Outcome {
	return { kind: 'failed', message: error instanceof Error ? error.message : '' };
}

export class EndingRecord {
	blockers = $state<PartnershipBlocker[]>([]);
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(partnershipId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.blockers = await listBlockers(this.#client(), partnershipId);
			this.status = 'ready';
		} catch {
			this.status = 'failed';
		}
	}

	async end(partnershipId: string, reason: string): Promise<Outcome> {
		this.busy = true;
		try {
			return {
				kind: 'done',
				partnership: await endPartnership(this.#client(), partnershipId, reason)
			};
		} catch (error) {
			await this.load(partnershipId);
			return failed(error);
		} finally {
			this.busy = false;
		}
	}

	async setListing(partnershipId: string, showOnPublic: boolean): Promise<Outcome> {
		this.busy = true;
		try {
			const partnership = await setPublicListing(this.#client(), partnershipId, showOnPublic);
			return { kind: 'done', partnership };
		} catch (error) {
			return failed(error);
		} finally {
			this.busy = false;
		}
	}
}
