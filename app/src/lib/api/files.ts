import type { ApiClient } from './client';
import type { components } from './schema';
import { shrinkImage } from './shrink-image';

export type StoredFile = components['schemas']['StoredFile'];
export type FilePurpose = StoredFile['purpose'];

export const pdfType = 'application/pdf';

export async function uploadImage(
	client: ApiClient,
	image: File,
	purpose: FilePurpose
): Promise<StoredFile> {
	const form = new FormData();
	form.set('purpose', purpose);
	if (image.type === pdfType) {
		form.set('file', image, 'bukti.pdf');
	} else {
		form.set('file', await shrinkImage(image), 'image');
	}
	return client.upload<StoredFile>('/api/v1/files', form);
}
