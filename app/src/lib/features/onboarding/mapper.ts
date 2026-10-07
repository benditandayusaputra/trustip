import type { FieldError } from '#lib/api/errors.js';
import type { OnboardingRequest } from '#lib/api/organizations.js';
import type { OnboardingErrors, OnboardingField, OnboardingValues } from './schema';

const fieldsFromServer: Record<string, OnboardingField | undefined> = {
	type: 'type',
	name: 'name',
	description: 'description',
	whatsapp: 'whatsapp',
	city_code: 'cityCode',
	district: 'district',
	address: 'address'
};

export function toOnboardingErrors(fields: FieldError[]): OnboardingErrors {
	const errors: OnboardingErrors = {};
	fields.forEach((problem) => {
		const field = fieldsFromServer[problem.field];
		if (field && errors[field] === undefined) {
			errors[field] = problem.message;
		}
	});
	return errors;
}

export function toOnboardingRequest(values: OnboardingValues): OnboardingRequest {
	return {
		type: values.type,
		name: values.name,
		description: values.description,
		whatsapp: values.whatsapp,
		city_code: values.cityCode,
		district: values.district,
		address: values.address
	};
}
