import { describe, expect, it } from 'vitest';
import { unzipStored } from '#lib/testing/unzip.js';
import { zipStored } from './zip';

const encode = (text: string) => new TextEncoder().encode(text);

describe('stored zip archive', () => {
	it('FR-RPT-08 keeps every entry readable in order', () => {
		const archive = zipStored([
			{ name: 'a.txt', data: encode('hello') },
			{ name: 'b/c.xml', data: encode('<x/>') }
		]);

		expect([...unzipStored(archive)]).toEqual([
			['a.txt', 'hello'],
			['b/c.xml', '<x/>']
		]);
	});

	it('writes the CRC-32 of each entry and a central directory that counts them', () => {
		const archive = zipStored([
			{ name: 'a.txt', data: encode('hello') },
			{ name: 'b.txt', data: encode('123456789') }
		]);
		const view = new DataView(archive.buffer);
		const end = archive.length - 22;

		expect(view.getUint32(14, true)).toBe(0x3610a686);
		expect(view.getUint32(end, true)).toBe(0x06054b50);
		expect(view.getUint16(end + 10, true)).toBe(2);
		expect(view.getUint32(view.getUint32(end + 16, true), true)).toBe(0x02014b50);
	});
});
