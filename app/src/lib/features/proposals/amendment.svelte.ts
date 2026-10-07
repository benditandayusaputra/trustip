import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { getAgreement, type Agreement } from '#lib/api/agreements.js';
import { listPartnerCatalog, type PartnerCatalogItem } from '#lib/api/partnerships.js';
import { getProduct } from '#lib/api/products.js';
import { sendAmendment, type InvitationRequest, type Proposal } from '#lib/api/proposals.js';
import type { SendOutcome } from './composer.svelte';
import {
	emptyInvitation,
	toInvitationRequest,
	validateInvitation,
	type InvitationErrors,
	type InvitationValues,
	type ItemValues
} from './schema';

const percentPerBasisPoint = 100;

type RowMode = 'ADD' | 'KEEP' | 'UPDATE' | 'STOP';

export interface AmendmentRow {
	mode: RowMode;
	item: ItemValues;
}

type AgreementItem = Agreement['items'][number];

function itemOf(product: AgreementItem): ItemValues {
	return {
		productId: product.product_id,
		name: product.product_name,
		unit: product.unit,
		model: product.pricing_model,
		retailPrice: product.retail_price,
		depositPrice: product.deposit_price,
		commissionPercent: String(product.commission_bps / percentPerBasisPoint)
	};
}

export function changedRows(rows: AmendmentRow[]): AmendmentRow[] {
	return rows.filter((row) => row.mode !== 'KEEP');
}

export function toAmendmentRequest(
	values: InvitationValues,
	rows: AmendmentRow[]
): InvitationRequest {
	const changed = changedRows(rows);
	const request = toInvitationRequest({ ...values, items: changed.map((row) => row.item) });
	const actions = { ADD: 'ADD', KEEP: 'ADD', UPDATE: 'UPDATE', STOP: 'STOP' } as const;
	request.items = request.items.map((item, index) => ({
		...item,
		action: actions[changed[index]?.mode ?? 'ADD']
	}));
	return request;
}

export class AmendmentComposer {
	values = $state<InvitationValues>(emptyInvitation());
	rows = $state<AmendmentRow[]>([]);
	addable = $state<PartnerCatalogItem[]>([]);
	errors = $state<InvitationErrors>({});
	status = $state<'loading' | 'ready' | 'failed'>('loading');
	failure = $state('');
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async load(partnershipId: string, prefillProductId: string): Promise<void> {
		try {
			const [agreement, catalog] = await Promise.all([
				getAgreement(this.#client(), partnershipId, 0),
				listPartnerCatalog(this.#client(), partnershipId)
			]);
			this.values = { ...emptyInvitation(), ...this.#termsOf(agreement) };
			this.rows = agreement.items
				.filter((product) => product.is_active)
				.map((product) => ({
					mode: 'KEEP',
					item: itemOf(product)
				}));
			this.addable = catalog.filter((product) => !product.terms || !product.is_active);
			await this.#prefill(prefillProductId);
			this.status = 'ready';
		} catch (error) {
			this.failure = error instanceof Error ? error.message : '';
			this.status = 'failed';
		}
	}

	#termsOf(agreement: Agreement): Partial<InvitationValues> {
		return {
			schedule: agreement.terms.settlement_schedule,
			paymentTermsDays: agreement.terms.payment_terms_days,
			rhythm: agreement.terms.count_rhythm,
			notes: agreement.terms.notes
		};
	}

	async #prefill(productId: string): Promise<void> {
		const candidate = this.addable.find((product) => product.product_id === productId);
		if (!candidate) {
			return;
		}
		const own = await getProduct(this.#client(), productId).catch(() => undefined);
		this.add(candidate, {
			retail: own?.suggested_retail_price,
			deposit: own?.suggested_deposit_price
		});
	}

	add(
		product: PartnerCatalogItem,
		suggested: { retail?: number | undefined; deposit?: number | undefined } = {}
	): void {
		const item: ItemValues = {
			productId: product.product_id,
			name: product.name,
			unit: product.unit,
			model: 'FIXED_DEPOSIT',
			retailPrice: suggested.retail,
			depositPrice: suggested.deposit,
			commissionPercent: '20'
		};
		this.rows = [...this.rows, { mode: 'ADD', item }];
		this.addable = this.addable.filter((candidate) => candidate.product_id !== product.product_id);
	}

	remove(productId: string): void {
		this.rows = this.rows.filter((row) => row.item.productId !== productId);
	}

	async send(partnershipId: string): Promise<SendOutcome> {
		this.errors = validateInvitation({
			...this.values,
			items: changedRows(this.rows).map((row) => row.item)
		});
		if (Object.keys(this.errors).length > 0) {
			return { kind: 'invalid' };
		}
		this.busy = true;
		try {
			const request = toAmendmentRequest(this.values, this.rows);
			const proposal: Proposal = await sendAmendment(this.#client(), partnershipId, request);
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
