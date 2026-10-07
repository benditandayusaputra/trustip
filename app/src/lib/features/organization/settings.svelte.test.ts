import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import SettingsHarness from '#lib/testing/SettingsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const organizationId = '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c';
const logoId = '0192f0c4-aa00-7e66-a1bb-4d2f3e5d6a90';
const tinyPng =
	'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==';

const store = {
	id: organizationId,
	type: 'STORE',
	name: 'Warung Bu Tini',
	slug: 'warung-bu-tini',
	description: '',
	whatsapp: '6281234567890',
	city_code: '33.74',
	city_name: 'Kota Semarang',
	district: 'Tembalang',
	address: '',
	location: null,
	logo: null
};

function envelope(data: unknown, status = 200): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } }, { status });
}

function list(data: unknown[]): Response {
	return Response.json({
		code: 'OK',
		message: '',
		data,
		meta: { request_id: 'r', next_cursor: null, has_more: false }
	});
}

type Route = (body: unknown) => Response;

function serve(routes: Record<string, Route>) {
	const sent: Record<string, unknown> = {};
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		sent[key] = typeof init.body === 'string' ? JSON.parse(init.body) : init.body;
		const all: Record<string, Route> = {
			'GET /api/v1/provinces': () => list([{ code: '33', name: 'Jawa Tengah' }]),
			'GET /api/v1/cities?province_code=33': () =>
				list([
					{
						code: '33.74',
						name: 'Kota Semarang',
						province_name: 'Jawa Tengah',
						latitude: -6.99,
						longitude: 110.42
					}
				]),
			[`GET /api/v1/organizations/${organizationId}`]: () => envelope(store),
			...routes
		};
		const respond = all[key];
		return Promise.resolve(
			respond
				? respond(sent[key])
				: Response.json(
						{ code: 'NOT_FOUND', message: 'x', meta: { request_id: 'r' } },
						{ status: 404 }
					)
		);
	});
	return sent;
}

afterEach(() => {
	vi.unstubAllGlobals();
});

function pngFile(): File {
	const bytes = Uint8Array.from(atob(tinyPng), (char) => char.charCodeAt(0));
	return new File([bytes], 'logo.png', { type: 'image/png' });
}

describe('OrganizationSettings', () => {
	it('FR-ORG-01 loads the profile, uploads a logo, and saves it together', async () => {
		const sent = serve({
			'POST /api/v1/files': () =>
				envelope({ id: logoId, url: 'http://localhost:8080/files/a.png' }, 201),
			[`PATCH /api/v1/organizations/${organizationId}`]: (body) =>
				envelope({ ...store, ...(body as object) })
		});
		const { container } = await render(SettingsHarness, { organizationId, canEdit: true });

		await expect.element(page.getByLabelText('Nama usaha')).toHaveValue('Warung Bu Tini');
		await expect.element(page.getByLabelText('Nomor WhatsApp usaha')).toHaveValue('081234567890');
		await expect.element(page.getByLabelText('Kabupaten atau kota')).toHaveValue('33.74');
		await expectAccessible(container);
		await userEvent.upload(page.getByLabelText('Pilih gambar logo'), pngFile());
		await expect.element(page.getByRole('img', { name: 'Logo Warung Bu Tini' })).toBeVisible();
		await userEvent.fill(page.getByLabelText('Nama usaha'), 'Warung Bu Tini Jaya');
		await userEvent.click(page.getByRole('button', { name: 'Simpan profil' }));

		await expect.element(page.getByText('Profil usaha tersimpan.')).toBeVisible();
		expect(sent[`PATCH /api/v1/organizations/${organizationId}`]).toMatchObject({
			name: 'Warung Bu Tini Jaya',
			city_code: '33.74',
			logo_file_id: logoId
		});
	});

	it('lets staff read the profile but not change it', async () => {
		serve({});
		await render(SettingsHarness, { organizationId, canEdit: false });

		await expect
			.element(page.getByText('Hanya pemilik usaha yang bisa mengubah profil.'))
			.toBeVisible();
		await expect.element(page.getByLabelText('Nama usaha')).toBeDisabled();
	});

	it('offers a retry when the profile cannot be loaded', async () => {
		serve({ [`GET /api/v1/organizations/${organizationId}`]: () => envelope(null, 503) });
		await render(SettingsHarness, { organizationId, canEdit: true });

		await expect.element(page.getByText('Profil usaha tidak bisa dimuat.')).toBeVisible();
	});
});
