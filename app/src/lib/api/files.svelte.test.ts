import { describe, expect, it } from 'vitest';
import { ApiClient } from './client';
import { uploadImage } from './files';

describe('uploadImage (R-SEC-54)', () => {
	it('sends the shrunk image under a fixed name together with its purpose', async () => {
		const sent: FormData[] = [];
		const client = new ApiClient({
			fetch: (_input, init) => {
				if (init?.body instanceof FormData) {
					sent.push(init.body);
				}
				return Promise.resolve(
					Response.json({ code: 'OK', message: '', data: { id: 'f1' }, meta: { request_id: 'r' } })
				);
			},
			newKey: () => 'key',
			withRefreshLock: (task) => task(),
			readCache: { remember: () => Promise.resolve(), recall: () => Promise.resolve(undefined) }
		});
		const canvas = new OffscreenCanvas(40, 30);
		canvas.getContext('2d')?.fillRect(0, 0, 40, 30);
		const photo = new File([await canvas.convertToBlob({ type: 'image/png' })], 'rumah-saya.png', {
			type: 'image/png'
		});

		await expect(uploadImage(client, photo, 'LOGO')).resolves.toEqual({ id: 'f1' });

		const file = sent[0]?.get('file');
		expect(sent[0]?.get('purpose')).toBe('LOGO');
		expect(file instanceof File && file.name).toBe('image');
		expect(file instanceof File && file.type).toBe('image/jpeg');
	});
});
