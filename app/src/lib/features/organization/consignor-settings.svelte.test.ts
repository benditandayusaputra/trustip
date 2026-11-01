import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import SettingsHarness from '#lib/testing/SettingsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const organizationId = '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c';
const qrId = '0192f0c4-aa00-7e66-a1bb-4d2f3e5d6a91';
const tinyPng =
	'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==';

const consignor = {
	id: organizationId,
	type: 'CONSIGNOR',
	name: 'Kripik Mbok Darmi',
	slug: 'kripik-mbok-darmi',
	description: '',
	whatsapp: '6281234567890',
	city_code: '33.74',
	city_name: 'Kota Semarang',
	district: '',
	address: '',
	location: null,
	logo: null,
	store_profile: null,
	consignor_profile: {
		delivery_radius_km: null,
		delivery_days: [],
		preferred_store_types: [],
		business_id_no: '',
		distribution_permit_no: '',
		halal_cert_no: '',
		can_provide_sample: false
	},
	payment: { info: '', qr: null }
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

function serve(routes: Record<string, () => Response>) {
	const sent: Record<string, unknown> = {};
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		sent[key] = typeof init.body === 'string' ? JSON.parse(init.body) : init.body;
		const all: Record<string, () => Response> = {
			'GET /api/v1/provinces': () => list([]),
			'GET /api/v1/cities?province_code=33': () => list([]),
			[`GET /api/v1/organizations/${organizationId}`]: () => envelope(consignor),
			...routes
		};
		return Promise.resolve(all[key]?.() ?? envelope(null, 404));
	});
	return sent;
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Consignor settings', () => {
	it('BR-ORG-07 marks legal numbers as declared by the owner and saves the profile', async () => {
		const path = `PATCH /api/v1/organizations/${organizationId}/consignor-profile`;
		const sent = serve({
			[path]: () => envelope(consignor),
			[`PATCH /api/v1/organizations/${organizationId}`]: () => envelope(consignor)
		});
		const { container } = await render(SettingsHarness, { organizationId, canEdit: true });

		await expect
			.element(
				page.getByText('Dinyatakan oleh pemilik usaha. Trustip tidak memverifikasinya.').first()
			)
			.toBeVisible();
		await userEvent.click(page.getByRole('button', { name: 'Edit profil' }));
		await userEvent.fill(page.getByLabelText('Jangkauan antar (km)', { exact: true }), '12');
		await userEvent.click(page.getByRole('checkbox', { name: 'Kamis' }));
		await userEvent.click(page.getByRole('checkbox', { name: 'Kafe atau kedai kopi' }));
		await userEvent.fill(page.getByLabelText('NIB'), '1234567890123');
		await userEvent.click(page.getByRole('switch', { name: 'Bisa mengirim sampel' }));
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Simpan profil', exact: true }));

		await expect.element(page.getByText('Profil usaha tersimpan.')).toBeVisible();
		expect(sent[path]).toEqual({
			delivery_radius_km: 12,
			delivery_days: ['THU'],
			preferred_store_types: ['CAFE'],
			business_id_no: '1234567890123',
			distribution_permit_no: '',
			halal_cert_no: '',
			can_provide_sample: true
		});
	});

	it('FR-ORG-07 uploads a private QRIS and saves it with the account text', async () => {
		const path = `PATCH /api/v1/organizations/${organizationId}/payment-info`;
		const sent = serve({
			'POST /api/v1/files': () => envelope({ id: qrId, url: `/api/v1/files/${qrId}` }, 201),
			[path]: () => envelope(consignor)
		});
		await render(SettingsHarness, { organizationId, canEdit: true, view: 'preferences' });

		await expect.element(page.getByLabelText('Rekening tujuan')).toBeDisabled();
		await userEvent.click(page.getByRole('button', { name: 'Edit', exact: true }));

		await userEvent.fill(page.getByLabelText('Rekening tujuan'), 'BCA 1234567 a.n. Darmi');
		const bytes = Uint8Array.from(atob(tinyPng), (char) => char.charCodeAt(0));
		await userEvent.upload(
			page.getByLabelText('Pilih gambar QRIS'),
			new File([bytes], 'qris.png', { type: 'image/png' })
		);
		await expect.element(page.getByRole('img', { name: 'QRIS pembayaran' })).toBeVisible();
		await userEvent.click(page.getByRole('button', { name: 'Simpan info pembayaran' }));

		await expect.element(page.getByText('Info pembayaran tersimpan.')).toBeVisible();
		await expect.element(page.getByLabelText('Rekening tujuan')).toBeDisabled();
		expect(sent[path]).toEqual({
			payment_info: 'BCA 1234567 a.n. Darmi',
			payment_qr_file_id: qrId
		});
	});
});
