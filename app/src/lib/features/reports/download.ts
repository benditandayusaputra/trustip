import { toXlsx, type Sheet } from '#lib/domain/xlsx.js';

const xlsxType = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet';

export function downloadXlsx(fileName: string, sheet: Sheet): void {
	const url = URL.createObjectURL(new Blob([toXlsx(sheet)], { type: xlsxType }));
	const link = document.createElement('a');
	link.href = url;
	link.download = fileName;
	link.click();
	URL.revokeObjectURL(url);
}
