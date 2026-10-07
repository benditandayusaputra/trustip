import { id } from '#lib/i18n/id.js';

export const steps = ['type', 'profile', 'location'] as const;
export type Step = (typeof steps)[number];

export interface GeoPoint {
	latitude: number;
	longitude: number;
}

export interface PlaceValues {
	provinceCode: string;
	cityCode: string;
	district: string;
	address: string;
	location: GeoPoint | null;
}

export interface ProfileValues extends PlaceValues {
	name: string;
	description: string;
	whatsapp: string;
}

export interface OnboardingValues extends ProfileValues {
	type: '' | 'CONSIGNOR' | 'STORE';
}

export type OnboardingField = keyof OnboardingValues;
export type OnboardingErrors = Partial<Record<OnboardingField, string>>;

interface TextRule {
	field: Exclude<OnboardingField, 'type' | 'provinceCode' | 'cityCode' | 'location'>;
	min: number;
	max: number;
}

const minNameLength = 2;
const maxNameLength = 80;
const maxDescriptionLength = 1000;
const maxDistrictLength = 80;
const maxAddressLength = 200;

const textRules: Record<Step, TextRule[]> = {
	type: [],
	profile: [
		{ field: 'name', min: minNameLength, max: maxNameLength },
		{ field: 'description', min: 0, max: maxDescriptionLength },
		{ field: 'whatsapp', min: 1, max: Infinity }
	],
	location: [
		{ field: 'district', min: 0, max: maxDistrictLength },
		{ field: 'address', min: 0, max: maxAddressLength }
	]
};

const choiceFields: Record<Step, OnboardingField[]> = {
	type: ['type'],
	profile: [],
	location: ['provinceCode', 'cityCode']
};

const locationFields: OnboardingField[] = ['location'];

export function emptyValues(): OnboardingValues {
	return {
		type: '',
		name: '',
		description: '',
		whatsapp: '',
		provinceCode: '',
		cityCode: '',
		district: '',
		address: '',
		location: null
	};
}

export function validateStep(step: Step, values: OnboardingValues): OnboardingErrors {
	const errors: OnboardingErrors = {};
	choiceFields[step]
		.filter((field) => values[field] === '')
		.forEach((field) => {
			errors[field] = id.onboarding.validation.choose;
		});
	textRules[step].forEach((rule) => {
		const problem = textProblem(values[rule.field], rule);
		if (problem) {
			errors[rule.field] = problem;
		}
	});
	return errors;
}

function textProblem(value: string, rule: TextRule): string | undefined {
	const length = Array.from(value.trim()).length;
	if (length === 0 && rule.min > 0) {
		return id.auth.validation.required;
	}
	if (length < rule.min) {
		return id.auth.validation.minLength(rule.min);
	}
	return length > rule.max ? id.onboarding.validation.maxLength(rule.max) : undefined;
}

export function validateProfile(values: OnboardingValues): OnboardingErrors {
	return { ...validateStep('profile', values), ...validateStep('location', values) };
}

export function validateAll(values: OnboardingValues): OnboardingErrors {
	return Object.assign({}, ...steps.map((step) => validateStep(step, values))) as OnboardingErrors;
}

export function stepOf(field: OnboardingField): Step {
	return (
		steps.find(
			(step) =>
				choiceFields[step].includes(field) || textRules[step].some((rule) => rule.field === field)
		) ?? (locationFields.includes(field) ? 'location' : 'type')
	);
}

export function hasErrors(errors: OnboardingErrors): boolean {
	return Object.keys(errors).length > 0;
}
