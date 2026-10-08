import { describe, expect, it } from 'vitest';
import { toCsv } from './csv';

describe('CSV export', () => {
	it('FR-RPT-08 writes a header and rows separated by commas with CRLF line ends', () => {
		expect(
			toCsv([
				['Kelompok', 'Omzet'],
				['Roti Manis', '36000']
			])
		).toBe('\uFEFFKelompok,Omzet\r\nRoti Manis,36000\r\n');
	});

	it.each(['=SUM(A1)', '+62812', '-1', '@cmd', '\tx', '\rx'])(
		'R-SEC-55 neutralises a cell that starts like a formula: %j',
		(cell) => {
			expect(toCsv([[cell]])).toBe(`\uFEFF${quoteIfNeeded(`'${cell}`)}\r\n`);
		}
	);

	it('quotes cells with commas, quotes, or line breaks', () => {
		expect(toCsv([['Toko "Maju", Jaya', 'a\nb']])).toBe('\uFEFF"Toko ""Maju"", Jaya","a\nb"\r\n');
	});
});

function quoteIfNeeded(cell: string): string {
	return /[",\r\n]/u.test(cell) ? `"${cell.replaceAll('"', '""')}"` : cell;
}
