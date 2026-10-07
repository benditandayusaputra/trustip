import type { ApiClient } from '#lib/api/client.js';
import { ApiError, NetworkError } from '#lib/api/errors.js';
import { registerUser } from '#lib/api/users.js';
import { toRegisterErrors, toRegisterRequest } from './mapper';
import {
	hasErrors,
	validateRegistration,
	type RegisterErrors,
	type RegisterValues
} from './schema';

const firstServerErrorStatus = 500;

export type SubmitOutcome =
	| { kind: 'created' }
	| { kind: 'invalid' }
	| { kind: 'failed'; message: string; canRetry: boolean };

export class RegisterForm {
	values = $state<RegisterValues>({
		fullName: '',
		email: '',
		password: '',
		whatsapp: '',
		privacyAccepted: false
	});
	errors = $state<RegisterErrors>({});
	captchaToken = $state('');
	busy = $state(false);
	readonly #client: () => ApiClient;

	constructor(client: () => ApiClient) {
		this.#client = client;
	}

	async submit(): Promise<SubmitOutcome> {
		this.errors = validateRegistration(this.values, this.captchaToken);
		if (hasErrors(this.errors)) {
			return { kind: 'invalid' };
		}
		this.busy = true;
		try {
			await registerUser(this.#client(), toRegisterRequest(this.values, this.captchaToken));
			return { kind: 'created' };
		} catch (error) {
			return this.#explain(error);
		} finally {
			this.busy = false;
			this.captchaToken = '';
		}
	}

	#explain(error: unknown): SubmitOutcome {
		if (error instanceof NetworkError) {
			return { kind: 'failed', message: error.message, canRetry: true };
		}
		if (!(error instanceof ApiError)) {
			throw error;
		}
		if (error.fields.length > 0) {
			this.errors = toRegisterErrors(error.fields);
			return { kind: 'invalid' };
		}
		if (error.code === 'CAPTCHA_FAILED') {
			this.errors = { captcha: error.message };
			return { kind: 'invalid' };
		}
		return {
			kind: 'failed',
			message: error.message,
			canRetry: error.status >= firstServerErrorStatus
		};
	}
}
