import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PartnersHarness from '#lib/testing/PartnersHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const partnershipId = '0192f0c4-0000-7000-8000-000000000501';

const summary = {
	id: partnershipId,
	status: 'ACTIVE',
	started_at: '2026-10-19T09:00:00Z',
	ended_at: null,
	counterpart: {
		id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c',
		name: 'Roti Pak Anton',
		type: 'CONSIGNOR'
	},
	stock_qty: 12,
	unpaid_amount: 14000
};

const partnership = {
	id: partnershipId,
	status: 'ACTIVE',
	started_at: '2026-10-19T09:00:00Z',
	ended_at: null,
	show_on_public: false,
	end_reason: '',
	counterpart: {
		...summary.counterpart,
		slug: 'roti-pak-anton',
		description: 'Roti harian.',
		whatsapp: '6281234567890',
		city_name: 'Kota Semarang',
		district: 'Tembalang',
		address: 'Jl. Mawar 3',
		location: { latitude: -6.99123, longitude: 110.42345 },
		logo_url: null
	}
};

const catalogItem = (name: string, terms: unknown) => ({
	product_id: `0192f0c4-0000-7000-8000-00000000010${name.length.toString()}`,
	name,
	sku: 'ROT-0001',
	unit: 'pcs',
	description: '',
	net_content: '',
	category_name: 'Roti dan kue basah',
	photo_url: null,
	is_active: true,
	terms
});

function list(data: unknown[]): Response {
	return Response.json({
		code: 'OK',
		message: '',
		data,
		meta: { request_id: 'r', next_cursor: null, has_more: false }
	});
}

function serve(routes: Record<string, () => Response>) {
	const calls: string[] = [];
	vi.stubGlobal('fetch', (path: string, init: RequestInit = {}) => {
		const method = init.method ?? 'GET';
		calls.push(method === 'GET' ? path : `${method} ${path}`);
		return Promise.resolve(
			(routes[`${init.method ?? 'GET'} ${path}`] ?? routes[path])?.() ??
				Response.json(
					{ code: 'NOT_FOUND', message: 'x', meta: { request_id: 'r' } },
					{ status: 404 }
				)
		);
	});
	return calls;
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Partners', () => {
	it('FR-PRT-01 lists active partners and switches to ended ones', async () => {
		const calls = serve({
			'/api/v1/partnerships?status=ACTIVE': () => list([summary]),
			'/api/v1/partnerships?status=ENDED': () => list([])
		});
		const { container } = await render(PartnersHarness, { view: 'list' });

		await expect.element(page.getByRole('link', { name: /Roti Pak Anton/u })).toBeVisible();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('tab', { name: 'Berakhir' }));

		await expect.element(page.getByText(/Belum ada mitra/u)).toBeVisible();
		expect(calls).toContain('/api/v1/partnerships?status=ENDED');
	});

	it('FR-PRT-02 shows the partner summary with a WhatsApp link', async () => {
		serve({
			[`/api/v1/partnerships/${partnershipId}`]: () =>
				Response.json({ code: 'OK', message: '', data: partnership, meta: { request_id: 'r' } }),
			[`/api/v1/partnerships/${partnershipId}/catalog`]: () => list([])
		});
		const { container } = await render(PartnersHarness, { view: 'page', partnershipId });
		await userEvent.click(page.getByText('Akhiri kemitraan', { exact: true }).first());

		await expect.element(page.getByText('Jl. Mawar 3')).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: /Hubungi lewat WhatsApp/u }))
			.toHaveAttribute('href', 'https://wa.me/6281234567890');
		await expectAccessible(container);
	});

	it('FR-PRD-06 shows the partner catalog with agreed terms on the agreement tab', async () => {
		const agreed = {
			pricing_model: 'FIXED_DEPOSIT',
			retail_price: 2000,
			commission_bps: 0,
			deposit_price: 1750,
			per_unit: { gross: 2000, store_share: 250, consignor_share: 1750 }
		};
		serve({
			[`/api/v1/partnerships/${partnershipId}`]: () =>
				Response.json({ code: 'OK', message: '', data: partnership, meta: { request_id: 'r' } }),
			[`/api/v1/partnerships/${partnershipId}/catalog`]: () =>
				list([catalogItem('Roti Manis', agreed), catalogItem('Keripik', null)])
		});
		await render(PartnersHarness, { view: 'page', partnershipId });

		await userEvent.click(page.getByRole('tab', { name: 'Kesepakatan' }));

		await expect.element(page.getByText('Kesepakatan tampil di sini')).toBeVisible();
		await expect.element(page.getByText('Belum ada di Kesepakatan')).toBeVisible();
		await expect.element(page.getByText('Rp1.750')).toBeVisible();
	});
});

describe('Partner agreement actions', () => {
	it('FR-AGR-06 offers the owner an amendment and a quick add for each new product', async () => {
		serve({
			[`/api/v1/partnerships/${partnershipId}`]: () =>
				Response.json({ code: 'OK', message: '', data: partnership, meta: { request_id: 'r' } }),
			[`/api/v1/partnerships/${partnershipId}/catalog`]: () => list([catalogItem('Keripik', null)])
		});
		await render(PartnersHarness, { view: 'page', partnershipId });

		await userEvent.click(page.getByRole('tab', { name: 'Kesepakatan' }));

		await expect
			.element(page.getByRole('link', { name: 'Ajukan perubahan' }))
			.toHaveAttribute('href', `/app/mitra/${partnershipId}/perubahan`);
		await expect
			.element(page.getByRole('link', { name: 'Usulkan ke toko' }))
			.toHaveAttribute(
				'href',
				`/app/mitra/${partnershipId}/perubahan?tambah=0192f0c4-0000-7000-8000-000000000107`
			);
	});
});

function envelope(data: unknown): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } });
}

const partnerPath = `/api/v1/partnerships/${partnershipId}`;

describe('Ending a partnership', () => {
	it('FR-PRT-01 shows the stock and unpaid balance of each partner', async () => {
		serve({ '/api/v1/partnerships?status=ACTIVE': () => list([summary]) });
		await render(PartnersHarness, { view: 'list' });

		await expect.element(page.getByText('Stok 12 unit · belum lunas Rp14.000')).toBeVisible();
	});

	it('BR-PRT-02 lists the blockers with links to their documents', async () => {
		serve({
			[partnerPath]: () => envelope(partnership),
			[`${partnerPath}/catalog`]: () => list([]),
			[`${partnerPath}/blockers`]: () =>
				list([
					{
						kind: 'UNPAID_SETTLEMENT',
						document_id: '0192f0c4-0000-7000-8000-000000000701',
						document_number: 'ST-202611-0001',
						qty: 0,
						amount: 14000,
						description: 'Setoran ST-202611-0001 belum lunas'
					}
				])
		});
		const { container } = await render(PartnersHarness, { view: 'page', partnershipId });
		await userEvent.click(page.getByText('Akhiri kemitraan', { exact: true }).first());

		await expect
			.element(page.getByRole('link', { name: 'Setoran ST-202611-0001 belum lunas' }))
			.toHaveAttribute('href', '/app/setoran/0192f0c4-0000-7000-8000-000000000701');
		await expectAccessible(container);
	});

	it('FR-PRT-04 ends the partnership after the owner confirms', async () => {
		const calls = serve({
			[partnerPath]: () => envelope(partnership),
			[`${partnerPath}/catalog`]: () => list([]),
			[`${partnerPath}/blockers`]: () => list([]),
			[`POST ${partnerPath}/end`]: () =>
				envelope({ ...partnership, status: 'ENDED', ended_at: '2026-11-20T03:00:00Z' })
		});
		await render(PartnersHarness, { view: 'page', partnershipId });

		await userEvent.click(page.getByText('Akhiri kemitraan', { exact: true }).first());
		await userEvent.click(
			page.getByRole('button', { name: 'Akhiri kemitraan', exact: true }).last()
		);
		await userEvent.click(
			page.getByRole('dialog').getByRole('button', { name: 'Akhiri kemitraan' })
		);

		await expect.element(page.getByText('Kemitraan berakhir.')).toBeVisible();
		expect(calls).toContain(`POST ${partnerPath}/end`);
	});

	it('FR-PRT-03 lets the store choose to be shown as a place to buy', async () => {
		const calls = serve({
			[partnerPath]: () => envelope(partnership),
			[`${partnerPath}/catalog`]: () => list([]),
			[`${partnerPath}/blockers`]: () => list([]),
			[`PATCH ${partnerPath}/public-listing`]: () =>
				envelope({ ...partnership, show_on_public: true })
		});
		await render(PartnersHarness, { view: 'page', partnershipId, isConsignor: false });

		await userEvent.click(
			page.getByRole('switch', {
				name: 'Tampilkan toko saya sebagai tempat membeli produk mitra ini'
			})
		);

		await expect.element(page.getByText('Pilihan tampil disimpan.')).toBeVisible();
		expect(calls).toContain(`PATCH ${partnerPath}/public-listing`);
	});
});
