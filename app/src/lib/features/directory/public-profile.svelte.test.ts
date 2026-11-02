import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import type { PublicConsignor, PublicNeed, PublicProduct, PublicStore } from '#lib/api/public.js';
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

const shop = {
	address: 'Jl. Sirojudin 12',
	district: 'Tembalang',
	city_name: 'Kota Semarang',
	profile_slug: 'warung-bu-tini',
	location: { latitude: -7.05, longitude: 110.43 }
};

const product: PublicProduct = {
	id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6b03',
	sku: 'ROT-0001',
	name: 'Roti Manis',
	unit: 'pcs',
	category: 'Roti dan kue basah',
	description: 'Lembut.',
	composition: 'Tepung, gula',
	net_content: '60 g',
	suggested_retail_price: 2000,
	photo_url: null,
	maker: {
		name: 'Roti Mak Ijah',
		district: 'Tembalang',
		city_name: 'Kota Semarang',
		profile_slug: null,
		legality: consignor.legality
	},
	stores: [
		{ ...shop, name: 'Kantin Kampus', distance_m: null },
		{ ...shop, name: 'Warung Bu Tini', distance_m: null }
	]
};

afterEach(() => {
	vi.unstubAllGlobals();
	vi.restoreAllMocks();
	Reflect.deleteProperty(navigator, 'share');
});

describe('Public profiles', () => {
	it('FR-PUB-02 shows the store terms, hours, open needs, and the newcomer mark', async () => {
		const { container } = await render(PublicProfileHarness, { store });

		await expect.element(page.getByRole('heading', { name: 'Warung Bu Tini' })).toBeVisible();
		await expect.element(page.getByText('Komisi 10% sampai 20%')).toBeVisible();
		await expect.element(page.getByText('Wajib punya izin edar')).toBeVisible();
		await expect.element(page.getByText('Butuh camilan kantin')).toBeVisible();
		await expect.element(page.getByText('Tepat bayar')).toBeVisible();
		await expect.element(page.getByText('Baru', { exact: true })).toBeVisible();
		await expect.element(page.getByText('Mitra aktif')).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-PUB-03 shows products with the suggested price and the declared legality', async () => {
		const { container } = await render(PublicProfileHarness, { consignor });

		await expect
			.element(page.getByRole('heading', { level: 1, name: 'Roti Mak Ijah' }))
			.toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Kembali ke daftar' }))
			.toHaveAttribute('href', '/cari/penitip?tab=penitip');
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

	it('FR-PUB-04 FR-PUB-05 shows the product and re-sorts the stores from the visitor position', async () => {
		vi.stubGlobal('fetch', (path: string) => {
			const nearest = {
				...product,
				stores: [
					{ ...shop, name: 'Warung Bu Tini', distance_m: 300 },
					{ ...shop, name: 'Kantin Kampus', distance_m: 2500 }
				]
			};
			return Promise.resolve(
				path.includes('lat=-7.05')
					? Response.json({ code: 'OK', message: '', data: nearest, meta: { request_id: 'r' } })
					: Response.json({ code: 'NOT_FOUND', meta: { request_id: 'r' } }, { status: 404 })
			);
		});
		const { container } = await render(PublicProfileHarness, {
			product,
			locate: () => Promise.resolve({ latitude: -7.05, longitude: 110.43 })
		});
		await expect.element(page.getByText('Tepung, gula')).toBeVisible();
		await expect.element(page.getByRole('heading', { level: 1, name: 'Roti Manis' })).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: /Roti Mak Ijah/u }))
			.toHaveAttribute('href', '/u/roti-mak-ijah');
		await expectAccessible(container);

		await userEvent.click(page.getByRole('button', { name: 'Urutkan dari lokasi saya' }));

		await expect.element(page.getByText('300 m', { exact: false })).toBeVisible();
		const first = page.getByRole('listitem').first();
		await expect.element(first).toHaveTextContent(/Warung Bu Tini/u);
	});

	it('BR-PUB-01 explains when no store shows the product', async () => {
		await render(PublicProfileHarness, { product: { ...product, stores: [] } });

		await expect
			.element(page.getByText('Belum ada toko yang menampilkan produk ini', { exact: false }))
			.toBeVisible();
	});

	it('keeps the name order when the visitor refuses the location', async () => {
		await render(PublicProfileHarness, { product });

		await userEvent.click(page.getByRole('button', { name: 'Urutkan dari lokasi saya' }));

		await expect
			.element(page.getByText('Lokasi tidak bisa dibaca', { exact: false }))
			.toBeVisible();
	});

	it('FR-PUB-06 FR-PUB-08 shows a shared need and copies its link', async () => {
		const need: PublicNeed = {
			id: '0192f0c4-0000-7000-8000-000000000301',
			title: 'Butuh camilan kantin',
			category: 'Kue kering dan camilan',
			description: 'Untuk kantin kampus.',
			est_qty_per_week: 50,
			max_retail_price: 5000,
			preferred_model: 'COMMISSION',
			offered_commission_bps: 2000,
			status: 'OPEN',
			expires_at: '2026-11-01T00:00:00Z',
			store: {
				id: store.id,
				name: 'Warung Bu Tini',
				store_type: 'CANTEEN',
				district: 'Tembalang',
				city_name: 'Kota Semarang',
				profile_slug: 'warung-bu-tini'
			}
		};
		const copied: string[] = [];
		Object.defineProperty(navigator, 'share', { value: undefined, configurable: true });
		vi.spyOn(navigator.clipboard, 'writeText').mockImplementation((text: string) => {
			copied.push(text);
			return Promise.resolve();
		});
		const { container } = await render(PublicProfileHarness, { need });
		await expect
			.element(page.getByRole('heading', { level: 1, name: 'Butuh camilan kantin' }))
			.toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Tanggapi dengan Ajuan' }))
			.toHaveAttribute('href', `/app/ajuan/baru?sasaran=${store.id}&kebutuhan=${need.id}`);
		await expectAccessible(container);

		await userEvent.click(page.getByRole('button', { name: 'Bagikan' }));

		await expect.element(page.getByText('Tautan disalin.')).toBeVisible();
		expect(copied).toEqual(['https://trustip.test/kebutuhan/x']);
	});
});
