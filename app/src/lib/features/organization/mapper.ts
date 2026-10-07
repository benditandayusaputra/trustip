import type { FieldError } from '#lib/api/errors.js';
import type {
	OnboardingRequest,
	Organization,
	ProfileUpdateRequest
} from '#lib/api/organizations.js';
import type { OnboardingErrors, OnboardingField, OnboardingValues, ProfileValues } from './schema';

const fieldsFromServer: Record<string, OnboardingField | undefined> = {
	type: 'type',
	name: 'name',
	description: 'description',
	whatsapp: 'whatsapp',
	city_code: 'cityCode',
	district: 'district',
	address: 'address',
	location: 'location'
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
		address: values.address,
		location: values.location
	};
}

export function toProfileValues(organization: Organization): ProfileValues {
	return {
		name: organization.name,
		description: organization.description,
		whatsapp: localWhatsApp(organization.whatsapp),
		provinceCode: organization.city_code.slice(0, provinceCodeLength),
		cityCode: organization.city_code,
		district: organization.district,
		address: organization.address,
		location: organization.location
	};
}

const provinceCodeLength = 2;
const countryCode = '62';

function localWhatsApp(normalized: string): string {
	return normalized.startsWith(countryCode)
		? `0${normalized.slice(countryCode.length)}`
		: normalized;
}

export function toProfileUpdateRequest(
	values: ProfileValues,
	logoFileId: string | null
): ProfileUpdateRequest {
	return {
		name: values.name,
		description: values.description,
		whatsapp: values.whatsapp,
		city_code: values.cityCode,
		district: values.district,
		address: values.address,
		location: values.location,
		logo_file_id: logoFileId
	};
}
