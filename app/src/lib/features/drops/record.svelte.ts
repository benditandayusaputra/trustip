import type { ApiClient } from '#lib/api/client.js';
import { cancelDrop, confirmDrop, getDrop, type Drop } from '#lib/api/drops.js';

export type DropChoice = 'confirm' | 'cancel';

function failureOf(error: unknown): string {
	return error instanceof Error ? error.message : '';
}

export class DropRecord {
	drop = $state<Drop>();
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(dropId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.drop = await getDrop(this.#client(), dropId);
			this.status = 'ready';
		} catch (error) {
			this.failure = failureOf(error);
			this.status = 'failed';
		}
	}

	async decide(choice: DropChoice): Promise<string | undefined> {
		const current = this.drop;
		if (!current) {
			return undefined;
		}
		const actions = { confirm: confirmDrop, cancel: cancelDrop };
		this.busy = true;
		try {
			this.drop = await actions[choice](this.#client(), current.id);
			return undefined;
		} catch (error) {
			return failureOf(error);
		} finally {
			this.busy = false;
		}
	}
}

export function canConfirm(drop: Drop, viewerOrganizationId: string): boolean {
	return drop.status === 'RECORDED' && drop.consignor.id === viewerOrganizationId;
}

export function canCancel(drop: Drop, viewerOrganizationId: string, now: Date): boolean {
	const isOpen = now.getTime() <= Date.parse(drop.cancel_deadline);
	return drop.status === 'RECORDED' && drop.store.id === viewerOrganizationId && isOpen;
}
