const locale = 'id-ID';

const groupedInteger = new Intl.NumberFormat(locale, { maximumFractionDigits: 0 });

export function formatInteger(value: number): string {
	return groupedInteger.format(value);
}

export function parseDigits(text: string): number | undefined {
	const digits = text.replace(/\D/gu, '');
	return digits === '' ? undefined : Number(digits);
}
