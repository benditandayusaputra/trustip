import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ConsignorDirectoryHarness from '#lib/testing/ConsignorDirectoryHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const traits = {
	legality: {
		has_distribution_permit: true,
		has_halal_certificate: false,
		can_provide_sample: true
	},
	delivery_radius_km: 10,
	reputation: { is_established: false, active_partners: 0 }
};

const bread = {
	id: '0192f0c4-0000-7000-8000-000000000201',
	name: 'Roti sobek',
	sku: 'RTS-0001',
	unit: 'pcs',
	category: { id: 'c1', name: 'Roti dan kue basah' },
	suggested_retail_price: 12000,
	shelf_life_days: 3,
	photo_url: null,
	consignor: { id: 'k1', slug: 'roti-bu-ani', name: 'Roti Bu Ani', city_name: 'Kota Semarang' },
	distance_m: 2000,
	traits,
	is_self: false,
	is_partner: false,
	match: { score: 62, label: 'GOOD', reasons: [{ code: 'DISTANCE', distance_m: 2000 }] }
};

const bakery = {
	id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6c02',
	slug: 'roti-bu-ani',
	name: 'Roti Bu Ani',
	city_name: 'Kota Semarang',
	district: 'Tembalang',
	logo_url: null,
	distance_m: 2000,
	categories: [{ id: 'c1', name: 'Roti dan kue basah' }],
	product_count: 2,
	joined_on: '2026-08-01',
	traits,
	is_self: false,
	is_partner: false,
	match: null
};

const search = {
	origin: { kind: 'ORGANIZATION', is_known: true },
	sort: 'NEAREST',
	selected_radius_km: 10,
	effective_radius_km: 10,
	is_expanded: false,
	category_ids: []
};

function list(data: unknown[]): Response {
	return Response.json({
		code: 'OK',
		message: '',
		data,
		meta: { request_id: 'r', next_cursor: null, has_more: false, search }
	});
}

function serve(routes: Record<string, () => Response>) {
	const requested: string[] = [];
	vi.stubGlobal('fetch', (path: string) => {
		requested.push(path);
		return Promise.resolve(routes[path]?.() ?? list([]));
	});
	return requested;
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Consignor directory', () => {
	it('FR-DIR-02 BR-DIR-01 opens on products with an approximate distance', async () => {
		serve({ '/api/v1/directory/products': () => list([bread]) });
		const { container } = await render(ConsignorDirectoryHarness, {});

		await expect.element(page.getByRole('heading', { name: 'Roti sobek' })).toBeVisible();
		await expect.element(page.getByText('± 2 km').first()).toBeVisible();
		await expect.element(page.getByText('Tahan 3 hari')).toBeVisible();
		await expect.element(page.getByText('Legalitas dinyatakan oleh pemilik usaha')).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-DIR-02 FR-DIR-12 switching to the consignor tab keeps it in the address', async () => {
		const requested = serve({ '/api/v1/directory/consignors': () => list([bakery]) });
		const addresses: string[] = [];
		await render(ConsignorDirectoryHarness, {
			onaddress: (address: string) => addresses.push(address)
		});

		await userEvent.click(page.getByRole('tab', { name: 'Penitip' }));

		await expect.element(page.getByText('2 produk', { exact: false })).toBeVisible();
		expect(requested).toContain('/api/v1/directory/consignors');
		expect(addresses).toContain('tab=penitip');
	});

	it('FR-DIR-07 asks again with the halal filter', async () => {
		const requested = serve({});
		await render(ConsignorDirectoryHarness, {});

		await userEvent.click(page.getByRole('button', { name: 'Filter dan urutan' }));
		await userEvent.click(page.getByRole('checkbox', { name: 'Punya sertifikat halal' }));

		await vi.waitFor(() => {
			expect(requested).toContain('/api/v1/directory/products?has_halal=true');
		});
	});
});
