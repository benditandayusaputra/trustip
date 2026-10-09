import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import type { PublicConsignor, PublicStore } from '#lib/api/public.js';
import PublicProfileHarness from '#lib/testing/PublicProfileHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const store: PublicStore = {
	id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6b01',
	slug: 'warung-bu-tini',
	name: 'Warung Bu Tini',
	store_type: 'GROCERY_STALL',
	description: 'Dekat SD.',
	address: 'Jl. Sirojudin 12',
	district: 'Tembalang',
	city_name: 'Kota Semarang',
	location: null,
	opening_hours: '07.00-21.00',
	logo_url: null,
	photo_urls: [],
	accepted_categories: [{ id: 'c1', name: 'Roti dan kue basah' }],
	terms: {
		accepts_fixed_deposit: false,
		accepts_commission: true,
		commission_min_bps: 1000,
		commission_max_bps: 2000,
		requires_distribution_permit: true,
		requires_halal: false,
		requires_sample: false
	},
	other_requirements: '',
	reputation: { is_new: true, on_time_bps: null, active_partners: 2 },
	joined_on: '2026-08-01',
	open_needs: [
		{
			id: '0192f0c4-0000-7000-8000-000000000301',
			title: 'Butuh camilan kantin',
			category: 'Kue kering dan camilan',
			expires_at: '2026-11-01T00:00:00Z'
		}
	]
};

const consignor: PublicConsignor = {
	id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6b02',
	slug: 'roti-mak-ijah',
	name: 'Roti Mak Ijah',
	description: 'Roti rumahan.',
	district: 'Tembalang',
	city_name: 'Kota Semarang',
	location: null,
	logo_url: null,
	legality: { business_id_no: '1234567890123', distribution_permit_no: null, halal_cert_no: null },
	active_partners: 3,
	joined_on: '2026-07-01',
	products: [
		{
			sku: 'ROT-0001',
			name: 'Roti Manis',
			unit: 'pcs',
			category: 'Roti dan kue basah',
			suggested_retail_price: 2000,
			photo_url: null
		}
	]
};

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Public profiles', () => {
	it('FR-PUB-02 shows the store terms, hours, open needs, and the newcomer mark', async () => {
		const { container } = await render(PublicProfileHarness, { store });

		await expect.element(page.getByRole('heading', { name: 'Warung Bu Tini' })).toBeVisible();
		await expect.element(page.getByText('Komisi 10% sampai 20%')).toBeVisible();
		await expect.element(page.getByText('Wajib punya izin edar')).toBeVisible();
		await expect.element(page.getByText('Butuh camilan kantin')).toBeVisible();
		await expect.element(page.getByText(/^Baru · 2 mitra aktif/u)).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-PUB-03 shows products with the suggested price and the declared legality', async () => {
		const { container } = await render(PublicProfileHarness, { consignor });

		await expect.element(page.getByText('Roti Manis')).toBeVisible();
		await expect.element(page.getByText('Rp2.000')).toBeVisible();
		await expect.element(page.getByText('1234567890123')).toBeVisible();
		await expect.element(page.getByText('3 toko mitra aktif', { exact: false })).toBeVisible();
		await expectAccessible(container);
	});

	it('asks a visitor without a session to sign in before showing the WhatsApp number', async () => {
		vi.stubGlobal('fetch', () =>
			Promise.resolve(
				Response.json(
					{ code: 'UNAUTHENTICATED', message: 'Silakan masuk.', meta: { request_id: 'r' } },
					{ status: 401 }
				)
			)
		);
		await render(PublicProfileHarness, { store });

		await userEvent.click(page.getByRole('button', { name: 'Tanya lewat WhatsApp' }));

		await expect
			.element(page.getByRole('link', { name: 'Masuk untuk melihat nomor WhatsApp' }))
			.toBeVisible();
	});
});
