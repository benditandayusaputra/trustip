import type { ApiClient } from './client';
import type { components } from './schema';
import { shrinkImage } from './shrink-image';

export type StoredFile = components['schemas']['StoredFile'];
export type FilePurpose = StoredFile['purpose'];

export async function uploadImage(
	client: ApiClient,
	image: File,
	purpose: FilePurpose
): Promise<StoredFile> {
	const form = new FormData();
	form.set('purpose', purpose);
	form.set('file', await shrinkImage(image), 'image');
	return client.upload<StoredFile>('/api/v1/files', form);
}
