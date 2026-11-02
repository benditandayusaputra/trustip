const localSignature = 0x04034b50;
const sizeOffset = 18;
const nameLengthOffset = 26;
const headerSize = 30;

export function unzipStored(bytes: Uint8Array): Map<string, string> {
	const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
	const decoder = new TextDecoder();
	const files = new Map<string, string>();
	let at = 0;
	while (at < bytes.length && view.getUint32(at, true) === localSignature) {
		const size = view.getUint32(at + sizeOffset, true);
		const start = at + headerSize + view.getUint16(at + nameLengthOffset, true);
		files.set(
			decoder.decode(bytes.subarray(at + headerSize, start)),
			decoder.decode(bytes.subarray(start, start + size))
		);
		at = start + size;
	}
	return files;
}
