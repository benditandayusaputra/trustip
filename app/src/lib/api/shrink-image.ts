export const maxImageWidth = 1280;
const bytesPerKibibyte = 1024;
export const maxImageBytes = bytesPerKibibyte * bytesPerKibibyte;

const acceptedTypes = ['image/jpeg', 'image/png', 'image/webp'];
const firstJpegQuality = 0.85;
const jpegQualityStep = 0.15;
const jpegAttempts = 4;
const jpegQualities = Array.from(
	{ length: jpegAttempts },
	(_, attempt) => firstJpegQuality - attempt * jpegQualityStep
);
const opaqueAlpha = 255;
const alphaOffset = 3;
const bytesPerPixel = 4;

class UnsupportedImageError extends Error {}

export async function shrinkImage(file: File): Promise<Blob> {
	if (!acceptedTypes.includes(file.type)) {
		throw new UnsupportedImageError(file.type);
	}
	const bitmap = await createImageBitmap(file, { imageOrientation: 'from-image' });
	const scale = Math.min(1, maxImageWidth / bitmap.width);
	const canvas = new OffscreenCanvas(
		Math.round(bitmap.width * scale),
		Math.round(bitmap.height * scale)
	);
	const context = canvas.getContext('2d');
	if (!context) {
		throw new UnsupportedImageError('canvas unavailable');
	}
	context.drawImage(bitmap, 0, 0, canvas.width, canvas.height);
	bitmap.close();
	return encodeWithinLimit(canvas, hasTransparency(context, canvas));
}

function hasTransparency(context: OffscreenCanvasRenderingContext2D, canvas: OffscreenCanvas) {
	const { data } = context.getImageData(0, 0, canvas.width, canvas.height);
	for (let index = alphaOffset; index < data.length; index += bytesPerPixel) {
		if (data[index] !== opaqueAlpha) {
			return true;
		}
	}
	return false;
}

async function encodeWithinLimit(canvas: OffscreenCanvas, isTransparent: boolean): Promise<Blob> {
	if (isTransparent) {
		const png = await canvas.convertToBlob({ type: 'image/png' });
		if (png.size <= maxImageBytes) {
			return png;
		}
	}
	for (const quality of jpegQualities) {
		const jpeg = await canvas.convertToBlob({ type: 'image/jpeg', quality });
		if (jpeg.size <= maxImageBytes) {
			return jpeg;
		}
	}
	throw new UnsupportedImageError('image stays above the size limit');
}
