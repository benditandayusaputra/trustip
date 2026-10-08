import { toCsv } from '#lib/domain/csv.js';

export function downloadCsv(fileName: string, rows: string[][]): void {
	const url = URL.createObjectURL(new Blob([toCsv(rows)], { type: 'text/csv;charset=utf-8' }));
	const link = document.createElement('a');
	link.href = url;
	link.download = fileName;
	link.click();
	URL.revokeObjectURL(url);
}
