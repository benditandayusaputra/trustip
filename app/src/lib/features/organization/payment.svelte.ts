import type { ApiClient } from '#lib/api/client.js';
import { ApiError } from '#lib/api/errors.js';
import { uploadImage } from '#lib/api/files.js';
import { updatePaymentInfo, type Organization } from '#lib/api/organizations.js';
import { organization as id } from '#lib/i18n/organization.js';
import { explainSectionFailure, type SectionOutcome } from './section';

type PaymentField = 'info' | 'qr';
type Payment = NonNullable<Organization['payment']>;

const fieldsFromServer: Record<string, PaymentField | undefined> = {
	payment_info: 'info',
	payment_qr_file_id: 'qr'
};

export class PaymentEditor {
	info = $state('');
	qr = $state<Payment['qr']>(null);
	errors = $state<Partial<Record<PaymentField, string>>>({});
	busy = $state(false);
	isUploading = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient, payment: Payment | null) {
		this.#client = client;
		this.info = payment?.info ?? '';
		this.qr = payment?.qr ?? null;
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
