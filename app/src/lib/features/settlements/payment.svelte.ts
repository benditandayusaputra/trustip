import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { uploadImage } from '#lib/api/files.js';
import {
	recordPayment,
	type PaymentMethod,
	type PaymentRequest,
	type Settlement
} from '#lib/api/settlements.js';
import { businessDateOf } from '#lib/i18n/format.js';
import { id } from '#lib/i18n/id.js';

type PaymentField = 'method' | 'paid_on' | 'proof_file_id' | 'amount';

export interface PaymentDeps {
	client: () => ApiClient;
	newKey: () => string;
	now: () => Date;
}

const proofMethods: PaymentMethod[] = ['TRANSFER', 'QRIS'];

export function needsProof(method: PaymentMethod): boolean {
	return proofMethods.includes(method);
}

export class PaymentComposer {
	method = $state<PaymentMethod>('TRANSFER');
	paidOn = $state('');
	proof = $state<{ id: string; url: string } | null>(null);
	errors = $state<Partial<Record<PaymentField, string>>>({});
	busy = $state(false);
	isUploading = $state(false);
	failure = $state('');
	#key: string;
	readonly #deps: PaymentDeps;

	constructor(deps: PaymentDeps) {
		this.#deps = deps;
		this.#key = deps.newKey();
		this.paidOn = this.today;
	}

	get today(): string {
		return businessDateOf(this.#deps.now());
	}

	async chooseProof(file: File): Promise<void> {
		this.isUploading = true;
		this.errors = {};
		try {
			const stored = await uploadImage(this.#deps.client(), file, 'PAYMENT_PROOF');
			this.proof = { id: stored.id, url: stored.url };
		} catch (error) {
			this.errors = {
				proof_file_id: error instanceof ApiError ? error.message : id.settings.imageInvalid
			};
		} finally {
			this.isUploading = false;
		}
	}

	request(settlement: Settlement): PaymentRequest {
		return {
			amount: settlement.split.consignor_share,
			method: this.method,
			paid_on: this.paidOn,
			proof_file_id: needsProof(this.method) ? (this.proof?.id ?? null) : null
		};
	}

	async submit(settlement: Settlement): Promise<Settlement | undefined> {
		if (needsProof(this.method) && !this.proof) {
			this.errors = { proof_file_id: id.settlements.proofRequired };
			return undefined;
		}
		this.busy = true;
		this.errors = {};
		this.failure = '';
		try {
			const saved = await recordPayment(this.#deps.client(), {
				settlementId: settlement.id,
				request: this.request(settlement),
				idempotencyKey: this.#key
			});
			this.#key = this.#deps.newKey();
			return saved;
		} catch (error) {
			this.#explain(error);
			return undefined;
		} finally {
			this.busy = false;
		}
	}

	#explain(error: unknown): void {
		if (!(error instanceof ApiError)) {
			this.failure = error instanceof Error ? error.message : '';
			return;
		}
		const errors: Partial<Record<PaymentField, string>> = {};
		for (const problem of error.fields) {
			errors[problem.field as PaymentField] = problem.message;
		}
		this.errors = errors;
		this.failure = error.fields.length > 0 ? '' : error.message;
	}
}
