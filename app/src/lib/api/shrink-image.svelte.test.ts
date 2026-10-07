import { describe, expect, it } from 'vitest';
import { maxImageBytes, maxImageWidth, shrinkImage } from './shrink-image';

async function canvasFile(
	width: number,
	height: number,
	options: { type: string; transparent?: boolean }
): Promise<File> {
	const canvas = new OffscreenCanvas(width, height);
	const context = canvas.getContext('2d');
	if (!context) {
		throw new Error('no 2d context');
	}
	if (!options.transparent) {
		context.fillStyle = 'gray';
		context.fillRect(0, 0, width, height);
	}
	context.fillStyle = 'white';
	context.fillRect(0, 0, width / 2, height / 2);
	const blob = await canvas.convertToBlob({ type: options.type });
	return new File([blob], 'foto-rumah.png', { type: options.type });
}

async function dimensionsOf(blob: Blob): Promise<[number, number]> {
	const bitmap = await createImageBitmap(blob);
	return [bitmap.width, bitmap.height];
}

describe('shrinkImage (BR-ORG-06)', () => {
	it('scales a wide photo down to 1280 pixels and re-encodes it as JPEG', async () => {
		const photo = await canvasFile(2560, 1440, { type: 'image/png' });

		const shrunk = await shrinkImage(photo);

		expect(shrunk.type).toBe('image/jpeg');
		expect(await dimensionsOf(shrunk)).toEqual([maxImageWidth, 720]);
		expect(shrunk.size).toBeLessThanOrEqual(maxImageBytes);
	});

	it('keeps a small image at its own size', async () => {
		const photo = await canvasFile(640, 480, { type: 'image/webp' });

		expect(await dimensionsOf(await shrinkImage(photo))).toEqual([640, 480]);
	});

	it('keeps transparency by encoding a see-through logo as PNG', async () => {
		const logo = await canvasFile(200, 200, { type: 'image/png', transparent: true });

		expect((await shrinkImage(logo)).type).toBe('image/png');
	});

	it('refuses files that are not JPG, PNG, or WebP', async () => {
		const svg = new File(['<svg onload="alert(1)"/>'], 'logo.svg', { type: 'image/svg+xml' });

		await expect(shrinkImage(svg)).rejects.toThrow();
	});
});
