import type { ApiClient } from '#lib/api/client.js';
import { getPartnership } from '#lib/api/partnerships.js';
import {
	confirmPayment,
	getSettlement,
	markSettlementPaid,
	rejectPayment,
	type Settlement
} from '#lib/api/settlements.js';

export type PaymentChoice = 'confirm' | 'reject' | 'markPaid';

export class SettlementRecord {
	settlement = $state<Settlement>();
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	busy = $state(false);
	counterpartWhatsApp = $state('');
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(settlementId: string): Promise<void> {
		this.status = 'loading';
		try {
			this.settlement = await getSettlement(this.#client(), settlementId);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	async loadCounterpart(partnershipId: string): Promise<void> {
		const partnership = await getPartnership(this.#client(), partnershipId).catch(() => undefined);
		this.counterpartWhatsApp = partnership?.counterpart.whatsapp ?? '';
	}

	replace(settlement: Settlement): void {
		this.settlement = settlement;
	}

	async decide(choice: PaymentChoice, reason = ''): Promise<string | undefined> {
		const current = this.settlement;
		if (!current) {
			return undefined;
		}
		const pending = current.payments.find((payment) => payment.status === 'PENDING_CONFIRMATION');
		const actions: Record<PaymentChoice, () => Promise<Settlement>> = {
			confirm: () => confirmPayment(this.#client(), pending?.id ?? ''),
			reject: () => rejectPayment(this.#client(), pending?.id ?? '', reason),
			markPaid: () => markSettlementPaid(this.#client(), current.id)
		};
		this.busy = true;
		try {
			this.settlement = await actions[choice]();
			return undefined;
		} catch (error) {
			return error instanceof Error ? error.message : '';
		} finally {
			this.busy = false;
		}
	}
}
