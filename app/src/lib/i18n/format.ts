const locale = 'id-ID';
const timeZone = 'Asia/Jakarta';
const metersPerKilometer = 1000;
const wholeKilometersFrom = 10;

const groupedInteger = new Intl.NumberFormat(locale, { maximumFractionDigits: 0 });
const kilometers = new Intl.NumberFormat(locale, {
	style: 'unit',
	unit: 'kilometer',
	maximumFractionDigits: 1
});
const wholeKilometers = new Intl.NumberFormat(locale, {
	style: 'unit',
	unit: 'kilometer',
	maximumFractionDigits: 0
});
const meters = new Intl.NumberFormat(locale, {
	style: 'unit',
	unit: 'meter',
	maximumFractionDigits: 0
});
const businessDate = new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeZone: 'UTC' });
const dateTime = new Intl.DateTimeFormat(locale, {
	dateStyle: 'medium',
	timeStyle: 'short',
	timeZone
});

export function formatInteger(value: number): string {
	return groupedInteger.format(value);
}

export function formatRupiah(amount: number): string {
	const sign = amount < 0 ? '-' : '';
	return `${sign}Rp${formatInteger(Math.abs(amount))}`;
}

export function formatBusinessDate(isoDate: string): string {
	return businessDate.format(new Date(`${isoDate}T00:00:00Z`));
}

export function formatDateTime(instant: string): string {
	return dateTime.format(new Date(instant));
}

export function formatDistance(distanceMeters: number): string {
	if (distanceMeters < metersPerKilometer) {
		return meters.format(distanceMeters);
	}
	const distanceKilometers = distanceMeters / metersPerKilometer;
	const format = distanceKilometers < wholeKilometersFrom ? kilometers : wholeKilometers;
	return format.format(distanceKilometers);
}

export function parseDigits(text: string): number | undefined {
	const digits = text.replace(/\D/gu, '');
	return digits === '' ? undefined : Number(digits);
}
