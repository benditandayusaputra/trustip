import type { ApiClient } from '#lib/api/client.js';
import { ApiError, NetworkError } from '#lib/api/errors.js';
import { logIn } from '#lib/api/sessions.js';
import { common as id } from '#lib/i18n/common.js';
import { deviceReadCache } from '#lib/offline/read-cache.js';

const firstServerErrorStatus = 500;
const formLevelCodes = new Set(['LOGIN_FAILED', 'ACCOUNT_LOCKED', 'RATE_LIMITED']);

export interface LoginValues {
	email: string;
	password: string;
}

export type LoginErrors = Partial<Record<keyof LoginValues | 'captcha' | 'form', string>>;

export type LoginOutcome =
	| { kind: 'signedIn' }
	| { kind: 'invalid' }
	| { kind: 'failed'; message: string; canRetry: boolean };

export class LoginForm {
	values = $state<LoginValues>({ email: '', password: '' });
	errors = $state<LoginErrors>({});
	captchaToken = $state('');
	busy = $state(false);
	readonly #client: () => ApiClient;
	readonly #forgetDeviceData: () => Promise<void>;

	constructor(client: () => ApiClient, forgetDeviceData = () => deviceReadCache().forget()) {
		this.#client = client;
		this.#forgetDeviceData = forgetDeviceData;
	}

	async submit(): Promise<LoginOutcome> {
		this.errors = validateLogin(this.values, this.captchaToken);
		if (Object.keys(this.errors).length > 0) {
			return { kind: 'invalid' };
		}
		this.busy = true;
		try {
			await logIn(this.#client(), {
				email: this.values.email,
				password: this.values.password,
				captcha_token: this.captchaToken
			});
			await this.#forgetDeviceData();
			return { kind: 'signedIn' };
		} catch (error) {
			return this.#explain(error);
		} finally {
			this.busy = false;
			this.captchaToken = '';
		}
	}

	#explain(error: unknown): LoginOutcome {
		if (error instanceof NetworkError) {
			return { kind: 'failed', message: error.message, canRetry: true };
		}
		if (!(error instanceof ApiError)) {
			throw error;
		}
		const explained = explainApiError(error);
		if (explained) {
			this.errors = explained;
			return { kind: 'invalid' };
		}
		return {
			kind: 'failed',
			message: error.message,
			canRetry: error.status >= firstServerErrorStatus
		};
	}
}

function explainApiError(error: ApiError): LoginErrors | undefined {
	if (formLevelCodes.has(error.code)) {
		return { form: error.message };
	}
	if (error.code === 'CAPTCHA_FAILED') {
		return { captcha: error.message };
	}
	const [first] = error.fields;
	return first ? { email: first.message } : undefined;
}

function validateLogin(values: LoginValues, captchaToken: string): LoginErrors {
	const errors: LoginErrors = {};
	if (values.email.trim() === '') {
		errors.email = id.auth.validation.required;
	}
	if (values.password === '') {
		errors.password = id.auth.validation.required;
	}
	if (captchaToken === '') {
		errors.captcha = id.auth.validation.captcha;
	}
	return errors;
}
