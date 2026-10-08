const formulaStart = /^[=+\-@\t\r]/u;
const needsQuotes = /[",\r\n]/u;
const byteOrderMark = '\uFEFF';
const lineEnd = '\r\n';

function neutralise(cell: string): string {
	return formulaStart.test(cell) ? `'${cell}` : cell;
}

function quote(cell: string): string {
	return needsQuotes.test(cell) ? `"${cell.replaceAll('"', '""')}"` : cell;
}

export function toCsv(rows: string[][]): string {
	const lines = rows.map((row) => row.map((cell) => quote(neutralise(cell))).join(','));
	return `${byteOrderMark}${lines.join(lineEnd)}${lineEnd}`;
}
