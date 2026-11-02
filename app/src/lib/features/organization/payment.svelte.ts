import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { uploadImage } from '#lib/api/files.js';
import { updatePaymentInfo, type Organization } from '#lib/api/organizations.js';
import { organization as id } from '#lib/i18n/organization.js';
import { explainSectionFailure, type SectionOutcome } from './section';

type Payment = NonNullable<Organization['payment']>;
export type PaymentMethod = Payment['methods'][number];
type PaymentField =
	'methods' | 'bankName' | 'bankAccountNumber' | 'bankAccountHolder' | 'info' | 'qr';

export const paymentMethods: PaymentMethod[] = ['CASH', 'TRANSFER', 'QRIS'];

const cashOnly: Payment = {
	info: '',
	qr: null,
	methods: ['CASH'],
	bank_name: '',
	bank_account_number: '',
	bank_account_holder: ''
};

const fieldsFromServer: Record<string, PaymentField | undefined> = {
	payment_methods: 'methods',
	bank_name: 'bankName',
	bank_account_number: 'bankAccountNumber',
	bank_account_holder: 'bankAccountHolder',
	payment_info: 'info',
	payment_qr_file_id: 'qr'
};

export class PaymentEditor {
	methods = $state<PaymentMethod[]>(['CASH']);
	bankName = $state('');
	bankAccountNumber = $state('');
	bankAccountHolder = $state('');
	info = $state('');
	qr = $state<Payment['qr']>(null);
	errors = $state<Partial<Record<PaymentField, string>>>({});
	busy = $state(false);
	isUploading = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient, payment: Payment | null) {
		this.#client = client;
		this.reset(payment);
	}

	reset(payment: Payment | null): void {
		const shown = payment ?? cashOnly;
		this.methods = shown.methods;
		this.bankName = shown.bank_name;
		this.bankAccountNumber = shown.bank_account_number;
		this.bankAccountHolder = shown.bank_account_holder;
		this.info = shown.info;
		this.qr = shown.qr;
		this.errors = {};
	}

	accepts(method: PaymentMethod): boolean {
		return this.methods.includes(method);
	}

	toggle(method: PaymentMethod, accepted: boolean): void {
		this.methods = paymentMethods.filter((candidate) =>
			candidate === method ? accepted : this.methods.includes(candidate)
		);
	}

	async chooseQr(file: File): Promise<void> {
		this.isUploading = true;
		this.errors = {};
		try {
			const stored = await uploadImage(this.#client(), file, 'PAYMENT_QR');
			this.qr = { id: stored.id, url: stored.url };
		} catch (error) {
			this.errors = { qr: error instanceof ApiError ? error.message : id.settings.imageInvalid };
		} finally {
			this.isUploading = false;
		}
	}

	async save(organizationId: string): Promise<SectionOutcome> {
		this.busy = true;
		this.errors = {};
		try {
			const saved = await updatePaymentInfo(this.#client(), organizationId, {
				payment_methods: this.methods,
				bank_name: this.bankName,
				bank_account_number: this.bankAccountNumber,
				bank_account_holder: this.bankAccountHolder,
				payment_info: this.info,
				payment_qr_file_id: this.qr?.id ?? null
			});
			return { kind: 'saved', organization: saved };
		} catch (error) {
			const failure = explainSectionFailure(error, fieldsFromServer);
			this.errors = failure.errors;
			return failure.outcome;
		} finally {
			this.busy = false;
		}
	}
}
