import { common as id } from '#lib/i18n/common.js';

const minFullNameLength = 2;
const minPasswordLength = 10;

export interface RegisterValues {
	fullName: string;
	email: string;
	password: string;
	whatsapp: string;
	privacyAccepted: boolean;
}

export type RegisterField = keyof RegisterValues | 'captcha';
export type RegisterErrors = Partial<Record<RegisterField, string>>;

const textRules: { field: Exclude<RegisterField, 'privacyAccepted' | 'captcha'>; min: number }[] = [
	{ field: 'fullName', min: minFullNameLength },
	{ field: 'email', min: 1 },
	{ field: 'password', min: minPasswordLength },
	{ field: 'whatsapp', min: 1 }
];

export function validateRegistration(values: RegisterValues, captchaToken: string): RegisterErrors {
	const errors: RegisterErrors = {};
	textRules.forEach(({ field, min }) => {
		const length = Array.from(values[field].trim()).length;
		if (length === 0) {
			errors[field] = id.auth.validation.required;
		} else if (length < min) {
			errors[field] = id.auth.validation.minLength(min);
		}
	});
	if (!values.privacyAccepted) {
		errors.privacyAccepted = id.auth.validation.consent;
	}
	if (captchaToken === '') {
		errors.captcha = id.auth.validation.captcha;
	}
	return errors;
}

export function hasErrors(errors: RegisterErrors): boolean {
	return Object.keys(errors).length > 0;
}
