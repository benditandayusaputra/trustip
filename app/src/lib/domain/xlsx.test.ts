import { describe, expect, it } from 'vitest';
import { unzipStored } from '#lib/testing/unzip.js';
import { toXlsx, type Sheet } from './xlsx';

function sheetOf(rows: Sheet['rows']): Map<string, string> {
	return unzipStored(toXlsx({ name: 'Penjualan', columnWidths: [30, 14], frozenRows: 1, rows }));
}

describe('Excel workbook', () => {
	it('FR-RPT-08 packs one styled worksheet with every part Excel needs', () => {
		const files = sheetOf([
			[
				{ value: 'Kelompok', style: 'header' },
				{ value: 'Omzet', style: 'header' }
			],
			[
				{ value: 'Roti Manis', style: 'text' },
				{ value: 36000, style: 'rupiah' }
			]
		]);
		const sheet = files.get('xl/worksheets/sheet1.xml') ?? '';

		expect([...files.keys()]).toEqual([
			'[Content_Types].xml',
			'_rels/.rels',
			'xl/workbook.xml',
			'xl/_rels/workbook.xml.rels',
			'xl/styles.xml',
			'xl/worksheets/sheet1.xml'
		]);
		expect(files.get('xl/workbook.xml')).toContain('<sheet name="Penjualan"');
		expect(files.get('xl/styles.xml')).toContain('formatCode="&quot;Rp&quot;#,##0"');
		expect(sheet).toContain('<col min="1" max="1" width="30" customWidth="1"/>');
		expect(sheet).toContain('<pane ySplit="1" topLeftCell="A2"');
		expect(sheet).toContain('<c r="B2" s="6"><v>36000</v></c>');
		expect(sheet).toContain(
			'<c r="A2" s="4" t="inlineStr"><is><t xml:space="preserve">Roti Manis</t>'
		);
	});

	it.each(['=SUM(A1)', '+62812', '-1', '@cmd'])(
		'R-SEC-55 stores formula-like text as plain text: %j',
		(value) => {
			const sheet = sheetOf([[{ value, style: 'text' }]]).get('xl/worksheets/sheet1.xml') ?? '';

			expect(sheet).toContain(`t="inlineStr"><is><t xml:space="preserve">${value}</t>`);
			expect(sheet).not.toContain('<f>');
		}
	);

	it('escapes markup and drops characters XML cannot hold', () => {
		const sheet =
			sheetOf([[{ value: 'Toko "Maju" & <Jaya>\u0007', style: 'text' }]]).get(
				'xl/worksheets/sheet1.xml'
			) ?? '';

		expect(sheet).toContain('Toko &quot;Maju&quot; &amp; &lt;Jaya&gt;</t>');
	});
});
