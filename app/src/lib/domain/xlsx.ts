import { zipStored } from './zip';

interface Look {
	format?: string;
	font?: string;
	fill?: string;
	border?: string;
	centered?: boolean;
}

const integerFormat = '3';
const percentFormat = '10';
const rupiahFormat = '164';
const boldFont = '1';
const titleFont = '2';
const headerFont = '3';
const noteFont = '4';
const headerFill = '2';
const totalFill = '3';
const thinBorder = '1';

const looks = {
	plain: {},
	title: { font: titleFont },
	note: { font: noteFont },
	header: { font: headerFont, fill: headerFill, border: thinBorder, centered: true },
	text: { border: thinBorder },
	integer: { format: integerFormat, border: thinBorder },
	rupiah: { format: rupiahFormat, border: thinBorder },
	percent: { format: percentFormat, border: thinBorder },
	totalText: { font: boldFont, fill: totalFill, border: thinBorder },
	totalInteger: { format: integerFormat, font: boldFont, fill: totalFill, border: thinBorder },
	totalRupiah: { format: rupiahFormat, font: boldFont, fill: totalFill, border: thinBorder },
	totalPercent: { format: percentFormat, font: boldFont, fill: totalFill, border: thinBorder }
} satisfies Record<string, Look>;

type CellStyle = keyof typeof looks;

export interface SheetCell {
	value: string | number;
	style: CellStyle;
}

export interface Sheet {
	name: string;
	columnWidths: number[];
	frozenRows: number;
	rows: SheetCell[][];
}

const styleNames = Object.keys(looks) as CellStyle[];
const xmlHead = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>';
const mainNs = 'http://schemas.openxmlformats.org/spreadsheetml/2006/main';
const packageRels = 'http://schemas.openxmlformats.org/package/2006/relationships';
const officeRels = 'http://schemas.openxmlformats.org/officeDocument/2006/relationships';
const officeType = 'application/vnd.openxmlformats-officedocument.spreadsheetml';
const firstLetter = 'A'.charCodeAt(0);
const alphabetSize = 26;
const disallowedXml = /[^\t\n\r\u0020-\u{10FFFF}]/gu;
const entities: Record<string, string> = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' };

function escapeXml(text: string): string {
	return text.replace(disallowedXml, '').replace(/[&<>"]/gu, (char) => entities[char] ?? char);
}

function columnName(index: number): string {
	const letter = String.fromCharCode(firstLetter + (index % alphabetSize));
	return index < alphabetSize ? letter : columnName(Math.floor(index / alphabetSize) - 1) + letter;
}

function xf(look: Look): string {
	const { format = '0', font = '0', fill = '0', border = '0' } = look;
	const alignment = look.centered
		? '<alignment horizontal="center" vertical="center" wrapText="1"/>'
		: '<alignment vertical="center"/>';
	return `<xf numFmtId="${format}" fontId="${font}" fillId="${fill}" borderId="${border}" xfId="0" applyNumberFormat="1" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1">${alignment}</xf>`;
}

const fonts = [
	'<font><sz val="11"/><name val="Calibri"/><family val="2"/></font>',
	'<font><b/><sz val="11"/><name val="Calibri"/><family val="2"/></font>',
	'<font><b/><sz val="16"/><color rgb="FF1E3A8A"/><name val="Calibri"/><family val="2"/></font>',
	'<font><b/><sz val="11"/><color rgb="FFFFFFFF"/><name val="Calibri"/><family val="2"/></font>',
	'<font><sz val="10"/><color rgb="FF64748B"/><name val="Calibri"/><family val="2"/></font>'
];
const fills = [
	'<fill><patternFill patternType="none"/></fill>',
	'<fill><patternFill patternType="gray125"/></fill>',
	'<fill><patternFill patternType="solid"><fgColor rgb="FF2563EB"/><bgColor indexed="64"/></patternFill></fill>',
	'<fill><patternFill patternType="solid"><fgColor rgb="FFE8EEFB"/><bgColor indexed="64"/></patternFill></fill>'
];
const thinSide = (side: string) => `<${side} style="thin"><color rgb="FFCBD5E1"/></${side}>`;
const borders = [
	'<border><left/><right/><top/><bottom/><diagonal/></border>',
	`<border>${['left', 'right', 'top', 'bottom'].map(thinSide).join('')}<diagonal/></border>`
];

function list(tag: string, items: string[]): string {
	return `<${tag} count="${String(items.length)}">${items.join('')}</${tag}>`;
}

const stylesXml = [
	xmlHead,
	`<styleSheet xmlns="${mainNs}">`,
	`<numFmts count="1"><numFmt numFmtId="${rupiahFormat}" formatCode="&quot;Rp&quot;#,##0"/></numFmts>`,
	list('fonts', fonts),
	list('fills', fills),
	list('borders', borders),
	'<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>',
	list(
		'cellXfs',
		styleNames.map((name) => xf(looks[name]))
	),
	'<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>',
	'</styleSheet>'
].join('');

function cellXml(cell: SheetCell, ref: string): string {
	const style = String(styleNames.indexOf(cell.style));
	if (typeof cell.value === 'number') {
		return `<c r="${ref}" s="${style}"><v>${String(cell.value)}</v></c>`;
	}
	return `<c r="${ref}" s="${style}" t="inlineStr"><is><t xml:space="preserve">${escapeXml(cell.value)}</t></is></c>`;
}

function rowXml(cells: SheetCell[], index: number): string {
	const row = String(index + 1);
	const content = cells.map((cell, column) => cellXml(cell, `${columnName(column)}${row}`));
	return `<row r="${row}">${content.join('')}</row>`;
}

function columnsXml(widths: number[]): string {
	const columns = widths.map((width, index) => {
		const position = String(index + 1);
		return `<col min="${position}" max="${position}" width="${String(width)}" customWidth="1"/>`;
	});
	return `<cols>${columns.join('')}</cols>`;
}

function sheetXml(sheet: Sheet): string {
	const split = String(sheet.frozenRows);
	const pane = `<pane ySplit="${split}" topLeftCell="A${String(sheet.frozenRows + 1)}" activePane="bottomLeft" state="frozen"/>`;
	return [
		xmlHead,
		`<worksheet xmlns="${mainNs}" xmlns:r="${officeRels}">`,
		'<sheetPr><pageSetUpPr fitToPage="1"/></sheetPr>',
		`<sheetViews><sheetView workbookViewId="0" showGridLines="0">${pane}</sheetView></sheetViews>`,
		'<sheetFormatPr defaultRowHeight="22" customHeight="1"/>',
		columnsXml(sheet.columnWidths),
		`<sheetData>${sheet.rows.map(rowXml).join('')}</sheetData>`,
		'<pageMargins left="0.5" right="0.5" top="0.75" bottom="0.75" header="0.3" footer="0.3"/>',
		'<pageSetup paperSize="9" orientation="landscape" fitToWidth="1" fitToHeight="0"/>',
		'</worksheet>'
	].join('');
}

const contentTypes = [
	xmlHead,
	'<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">',
	'<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>',
	'<Default Extension="xml" ContentType="application/xml"/>',
	`<Override PartName="/xl/workbook.xml" ContentType="${officeType}.sheet.main+xml"/>`,
	`<Override PartName="/xl/worksheets/sheet1.xml" ContentType="${officeType}.worksheet+xml"/>`,
	`<Override PartName="/xl/styles.xml" ContentType="${officeType}.styles+xml"/>`,
	'</Types>'
].join('');

const rootRels = `${xmlHead}<Relationships xmlns="${packageRels}"><Relationship Id="rId1" Type="${officeRels}/officeDocument" Target="xl/workbook.xml"/></Relationships>`;

const workbookRels = `${xmlHead}<Relationships xmlns="${packageRels}"><Relationship Id="rId1" Type="${officeRels}/worksheet" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Type="${officeRels}/styles" Target="styles.xml"/></Relationships>`;

function workbookXml(name: string): string {
	return `${xmlHead}<workbook xmlns="${mainNs}" xmlns:r="${officeRels}"><sheets><sheet name="${escapeXml(name)}" sheetId="1" r:id="rId1"/></sheets></workbook>`;
}

export function toXlsx(sheet: Sheet): Uint8Array<ArrayBuffer> {
	const encoder = new TextEncoder();
	const parts = [
		['[Content_Types].xml', contentTypes],
		['_rels/.rels', rootRels],
		['xl/workbook.xml', workbookXml(sheet.name)],
		['xl/_rels/workbook.xml.rels', workbookRels],
		['xl/styles.xml', stylesXml],
		['xl/worksheets/sheet1.xml', sheetXml(sheet)]
	] as const;
	return zipStored(parts.map(([name, xml]) => ({ name, data: encoder.encode(xml) })));
}
