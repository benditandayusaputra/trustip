import type { FieldError } from '#lib/api/errors.js';
import type { RegisterRequest } from '#lib/api/users.js';
import type { RegisterErrors, RegisterField, RegisterValues } from './schema';

const fieldsFromServer: Record<string, RegisterField | undefined> = {
	full_name: 'fullName',
	email: 'email',
	password: 'password',
	whatsapp: 'whatsapp',
	privacy_accepted: 'privacyAccepted',
	captcha_token: 'captcha'
};

export function toRegisterErrors(fields: FieldError[]): RegisterErrors {
	const errors: RegisterErrors = {};
	fields.forEach((problem) => {
		const field = fieldsFromServer[problem.field];
		if (field && errors[field] === undefined) {
			errors[field] = problem.message;
		}
	});
	return errors;
}

export function toRegisterRequest(values: RegisterValues, captchaToken: string): RegisterRequest {
	return {
		full_name: values.fullName,
		email: values.email,
		password: values.password,
		whatsapp: values.whatsapp,
		privacy_accepted: values.privacyAccepted,
		captcha_token: captchaToken
	};
}
