export interface ZipEntry {
	name: string;
	data: Uint8Array;
}

type Field = readonly [width: number, value: number];

const short = 2;
const long = 4;
const byteValues = 256;
const bitsPerByte = 8;
const crcPolynomial = 0xedb88320;
const crcMask = 0xffffffff;
const lowByte = 0xff;
const localSignature = 0x04034b50;
const centralSignature = 0x02014b50;
const endSignature = 0x06054b50;
const zipVersion = 20;
const utf8Names = 0x0800;
const stored = 0;
const dosEpochDate = 0x21;

const crcTable = Array.from({ length: byteValues }, (_, index) => {
	let value = index;
	for (let bit = 0; bit < bitsPerByte; bit++) {
		value = value & 1 ? crcPolynomial ^ (value >>> 1) : value >>> 1;
	}
	return value >>> 0;
});

function crc32(data: Uint8Array): number {
	let crc = crcMask;
	for (const byte of data) {
		crc = (crcTable[(crc ^ byte) & lowByte] ?? 0) ^ (crc >>> bitsPerByte);
	}
	return (crc ^ crcMask) >>> 0;
}

function concat(chunks: Uint8Array[]): Uint8Array<ArrayBuffer> {
	const bytes = new Uint8Array(chunks.reduce((sum, chunk) => sum + chunk.length, 0));
	chunks.reduce((at, chunk) => {
		bytes.set(chunk, at);
		return at + chunk.length;
	}, 0);
	return bytes;
}

function record(fields: Field[]): Uint8Array {
	const view = new DataView(new ArrayBuffer(fields.reduce((sum, [width]) => sum + width, 0)));
	fields.reduce((at, [width, value]) => {
		if (width === short) {
			view.setUint16(at, value, true);
		} else {
			view.setUint32(at, value, true);
		}
		return at + width;
	}, 0);
	return new Uint8Array(view.buffer);
}

function entryFields(name: Uint8Array, data: Uint8Array): Field[] {
	return [
		[short, zipVersion],
		[short, utf8Names],
		[short, stored],
		[short, 0],
		[short, dosEpochDate],
		[long, crc32(data)],
		[long, data.length],
		[long, data.length],
		[short, name.length],
		[short, 0]
	];
}

function entryRecords(entry: ZipEntry, offset: number) {
	const name = new TextEncoder().encode(entry.name);
	const fields = entryFields(name, entry.data);
	const header = record([
		[long, centralSignature],
		[short, zipVersion],
		...fields,
		[short, 0],
		[short, 0],
		[short, 0],
		[long, 0],
		[long, offset]
	]);
	return {
		local: concat([record([[long, localSignature], ...fields]), name, entry.data]),
		central: concat([header, name])
	};
}

function endRecord(count: number, directorySize: number, directoryOffset: number): Uint8Array {
	return record([
		[long, endSignature],
		[short, 0],
		[short, 0],
		[short, count],
		[short, count],
		[long, directorySize],
		[long, directoryOffset],
		[short, 0]
	]);
}

export function zipStored(entries: ZipEntry[]): Uint8Array<ArrayBuffer> {
	const locals: Uint8Array[] = [];
	const centrals: Uint8Array[] = [];
	let offset = 0;
	for (const entry of entries) {
		const { local, central } = entryRecords(entry, offset);
		locals.push(local);
		centrals.push(central);
		offset += local.length;
	}
	const directory = concat(centrals);
	return concat([...locals, directory, endRecord(entries.length, directory.length, offset)]);
}
