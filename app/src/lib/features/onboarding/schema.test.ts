import { describe, expect, it } from 'vitest';
import { emptyValues, stepOf, validateAll, validateStep, type OnboardingValues } from './schema';

function filled(): OnboardingValues {
	return {
		...emptyValues(),
		type: 'STORE',
		name: 'Warung Bu Tini',
		whatsapp: '0812 3456 7890',
		provinceCode: '33',
		cityCode: '33.74'
	};
}

describe('validateStep (FR-AUTH-04)', () => {
	it('asks for a business type on the first step', () => {
		expect(validateStep('type', emptyValues())).toEqual({ type: 'Pilih salah satu.' });
	});

	it('asks for name and WhatsApp on the profile step but not for the description', () => {
		expect(Object.keys(validateStep('profile', emptyValues()))).toEqual(['name', 'whatsapp']);
	});

	it('asks for province and city on the location step', () => {
		expect(Object.keys(validateStep('location', emptyValues()))).toEqual([
			'provinceCode',
			'cityCode'
		]);
	});

	it.each([
		[1, 'Minimal 2 karakter.'],
		[2, undefined],
		[80, undefined],
		[81, 'Maksimal 80 karakter.']
	])('R-SEC-46 a name of %i characters gives %s', (length, message) => {
		const values = { ...filled(), name: 'é'.repeat(length) };

		expect(validateStep('profile', values).name).toBe(message);
	});

	it('accepts a complete form', () => {
		expect(validateAll(filled())).toEqual({});
	});
});

describe('stepOf', () => {
	it.each([
		['type', 'type'],
		['whatsapp', 'profile'],
		['cityCode', 'location'],
		['address', 'location']
	] as const)('places %s on the %s step', (field, step) => {
		expect(stepOf(field)).toBe(step);
	});
});
