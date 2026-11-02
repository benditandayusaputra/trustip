import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { pdfType, uploadImage } from '#lib/api/files.js';
import {
	recordPayment,
	type PaymentMethod,
	type PaymentRequest,
	type Settlement
} from '#lib/api/settlements.js';
import { businessDateOf } from '#lib/i18n/format.js';
import { settlements as id } from '#lib/i18n/settlements.js';

type PaymentField = 'method' | 'paid_on' | 'proof_file_id' | 'amount';

export interface PaymentDeps {
	client: () => ApiClient;
	newKey: () => string;
	now: () => Date;
}

const allMethods: PaymentMethod[] = ['CASH', 'TRANSFER', 'QRIS'];

export function acceptedMethods(settlement: Settlement): PaymentMethod[] {
	const methods = settlement.payee?.methods ?? [];
	return methods.length > 0 ? methods : allMethods;
}

interface Proof {
	id: string;
	url: string;
	documentName: string | undefined;
}

export class PaymentComposer {
	method = $state<PaymentMethod>('CASH');
	paidOn = $state('');
	proof = $state<Proof | null>(null);
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
			this.proof = {
				id: stored.id,
				url: stored.url,
				documentName: stored.content_type === pdfType ? file.name : undefined
			};
		} catch (error) {
			this.errors = {
				proof_file_id: error instanceof ApiError ? error.message : id.settlements.proofInvalid
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
			proof_file_id: this.proof?.id ?? null
		};
	}

	async submit(settlement: Settlement): Promise<Settlement | undefined> {
		if (!this.proof) {
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
