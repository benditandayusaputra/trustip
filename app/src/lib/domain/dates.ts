const millisecondsPerDay = 86_400_000;
const isoDateLength = 10;

export function addDays(isoDate: string, days: number): string {
	return new Date(Date.parse(isoDate) + days * millisecondsPerDay)
		.toISOString()
		.slice(0, isoDateLength);
}

export function daysBetween(fromIsoDate: string, toIsoDate: string): number {
	return Math.round((Date.parse(toIsoDate) - Date.parse(fromIsoDate)) / millisecondsPerDay);
}
