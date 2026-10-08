import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import DirectoryHarness from '#lib/testing/DirectoryHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const storesPath = '/api/v1/directory/stores';

const canteen = {
	id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6b01',
	slug: 'kantin-kampus',
	name: 'Kantin Kampus',
	city_name: 'Kota Semarang',
	district: 'Tembalang',
	store_type: 'CANTEEN',
	logo_url: null,
	distance_m: 1200,
	accepted_categories: [{ id: 'c1', name: 'Kue kering dan camilan' }],
	terms: {
		accepts_fixed_deposit: false,
		accepts_commission: true,
		commission_min_bps: 1000,
		commission_max_bps: 2000,
		requires_distribution_permit: false,
		requires_halal: false,
		requires_sample: false
	},
	reputation: { is_new: false, on_time_bps: 9600 },
	is_self: false,
	is_partner: true,
	match: {
		score: 77,
		label: 'VERY_GOOD',
		reasons: [
			{ code: 'DISTANCE', distance_m: 1200 },
			{ code: 'ACCEPTS_CATEGORY', category: 'Kue kering dan camilan' },
			{ code: 'ON_TIME_PAYMENT', on_time_bps: 9600 }
		]
	}
};

function search(overrides: Record<string, unknown> = {}) {
	return {
		origin: { kind: 'ORGANIZATION', is_known: true },
		sort: 'NEAREST',
		selected_radius_km: 10,
		effective_radius_km: 10,
		is_expanded: false,
		category_ids: [],
		...overrides
	};
}

function list(data: unknown[], searched = search()): Response {
	return Response.json({
		code: 'OK',
		message: '',
		data,
		meta: { request_id: 'r', next_cursor: null, has_more: false, search: searched }
	});
}

function serve(routes: Record<string, () => Response>) {
	const requested: string[] = [];
	vi.stubGlobal('fetch', (path: string) => {
		requested.push(path);
		if (path === '/api/v1/categories') {
			return Promise.resolve(list([]));
		}
		return Promise.resolve(
			routes[path]?.() ??
				Response.json(
					{ code: 'INTERNAL', message: 'Terjadi kesalahan.', meta: { request_id: 'r' } },
					{ status: 500 }
				)
		);
	});
	return requested;
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Store directory', () => {
	it('FR-DIR-01 FR-DIR-10 shows the card with distance, score, and short reasons', async () => {
		serve({ [storesPath]: () => list([canteen]) });
		const { container } = await render(DirectoryHarness, {});

		await expect.element(page.getByRole('heading', { name: 'Kantin Kampus' })).toBeVisible();
		await expect.element(page.getByText('Sangat cocok')).toBeVisible();
		await expect.element(page.getByText('Bayar tepat waktu 96%').first()).toBeVisible();
		await expect.element(page.getByText('Terima Kue kering dan camilan')).toBeVisible();
		await expect.element(page.getByText('Sudah bermitra')).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-DIR-06 FR-DIR-12 choosing a store type asks again and updates the address', async () => {
		const requested = serve({
			[storesPath]: () => list([canteen]),
			[`${storesPath}?store_type=CANTEEN`]: () => list([canteen])
		});
		const addresses: string[] = [];
		await render(DirectoryHarness, { onaddress: (address: string) => addresses.push(address) });

		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Jenis toko' }), 'CANTEEN');

		await vi.waitFor(() => {
			expect(requested).toContain(`${storesPath}?store_type=CANTEEN`);
		});
		expect(addresses).toContain('store_type=CANTEEN');
	});

	it('BR-DIR-03 tells when the radius was widened', async () => {
		serve({
			[`${storesPath}?radius_km=2`]: () =>
				list(
					[canteen],
					search({ selected_radius_km: 2, effective_radius_km: 25, is_expanded: true })
				)
		});
		await render(DirectoryHarness, { address: 'radius_km=2' });

		await expect.element(page.getByText('Diperluas ke 25 km')).toBeVisible();
	});

	it('BR-DIR-04 asks for a location when the organization has none', async () => {
		serve({
			[storesPath]: () => list([], search({ origin: { kind: 'ORGANIZATION', is_known: false } }))
		});
		await render(DirectoryHarness, {});

		await expect.element(page.getByRole('link', { name: 'Atur lokasi usaha' })).toBeVisible();
		await expect
			.element(page.getByText('Belum ada toko yang cocok dengan saringan ini.'))
			.toBeVisible();
	});

	it('keeps the filters and offers a retry when loading fails', async () => {
		serve({});
		await render(DirectoryHarness, { address: 'store_type=CAFE' });

		await expect.element(page.getByText('Daftar toko tidak bisa dimuat.')).toBeVisible();
		await expect.element(page.getByRole('button', { name: 'Coba lagi' })).toBeVisible();
		await expect.element(page.getByRole('combobox', { name: 'Jenis toko' })).toHaveValue('CAFE');
	});
});
