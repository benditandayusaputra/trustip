import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ProductsHarness from '#lib/testing/ProductsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const breadCategory = '01a1161e-0a30-752d-ab9b-7397976f0dba';
const productId = '0192f0c4-0000-7000-8000-000000000101';

const bread = {
	id: productId,
	sku: 'ROT-0001',
	name: 'Roti Sobek Cokelat',
	category_id: breadCategory,
	category_name: 'Roti dan kue basah',
	unit: 'pcs',
	description: '',
	composition: '',
	net_content: '60 g',
	photo: null,
	suggested_retail_price: 2000,
	suggested_deposit_price: 1750,
	shelf_life_days: 3,
	status: 'ACTIVE',
	created_at: '2026-10-19T08:00:00Z',
	updated_at: '2026-10-19T08:00:00Z'
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

const categories = () => list([{ id: breadCategory, name: 'Roti dan kue basah', slug: 'roti' }]);

function serve(routes: Record<string, () => Response>) {
	const sent: Record<string, unknown> = {};
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		sent[key] = typeof init.body === 'string' ? JSON.parse(init.body) : init.body;
		const route = routes[key] ?? routes[key.split('?')[0] ?? ''];
		return Promise.resolve(
			route?.() ??
				Response.json(
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

describe('Product catalog', () => {
	it('FR-PRD-05 lists products and searches by name after a pause', async () => {
		const sent = serve({
			'GET /api/v1/categories': categories,
			'GET /api/v1/products?status=ACTIVE': () => list([bread]),
			'GET /api/v1/products?status=ACTIVE&q=roti': () => list([bread])
		});
		const { container } = await render(ProductsHarness, { view: 'catalog' });

		await expect.element(page.getByRole('link', { name: /Roti Sobek Cokelat/u })).toBeVisible();
		await userEvent.type(page.getByRole('searchbox', { name: 'Cari nama atau SKU' }), 'roti');

		await vi.waitFor(() => {
			expect(Object.keys(sent)).toContain('GET /api/v1/products?status=ACTIVE&q=roti');
		});
		await expectAccessible(container);
	});

	it('shows an empty state that leads to the product form', async () => {
		serve({
			'GET /api/v1/categories': categories,
			'GET /api/v1/products?status=ACTIVE': () => list([])
		});
		await render(ProductsHarness, { view: 'catalog' });

		await expect
			.element(page.getByRole('link', { name: 'Tambah produk' }))
			.toHaveAttribute('href', '/app/produk/baru');
	});
});

describe('Product form', () => {
	it('BR-AGR-02 stops a deposit price above the retail price before sending', async () => {
		const sent = serve({ 'GET /api/v1/categories': categories });
		await render(ProductsHarness, { view: 'form' });

		await userEvent.type(page.getByLabelText('Nama produk'), 'Roti Sobek');
		await userEvent.selectOptions(page.getByLabelText('Kategori'), breadCategory);
		await userEvent.type(page.getByLabelText('Harga jual saran'), '2000');
		await userEvent.type(page.getByLabelText('Harga setor saran'), '2500');
		await userEvent.click(page.getByRole('button', { name: 'Simpan produk' }));

		await expect
			.element(page.getByText('Harga setor tidak boleh melebihi harga jual.'))
			.toBeVisible();
		expect(Object.keys(sent)).not.toContain('POST /api/v1/products');
	});

	it('FR-PRD-02 sends an empty SKU so the server generates it', async () => {
		const sent = serve({
			'GET /api/v1/categories': categories,
			'POST /api/v1/products': () => envelope(bread, 201)
		});
		const onsaved = vi.fn();
		await render(ProductsHarness, { view: 'form', onsaved });

		await userEvent.type(page.getByLabelText('Nama produk'), 'Roti Sobek Cokelat');
		await userEvent.selectOptions(page.getByLabelText('Kategori'), breadCategory);
		await userEvent.type(page.getByLabelText('Harga jual saran'), '2000');
		await userEvent.type(page.getByLabelText('Harga setor saran'), '1750');
		await userEvent.click(page.getByRole('button', { name: 'Simpan produk' }));

		await expect.element(page.getByText('Produk tersimpan.')).toBeVisible();
		expect(sent['POST /api/v1/products']).toMatchObject({
			sku: '',
			suggested_retail_price: 2000,
			suggested_deposit_price: 1750,
			shelf_life_days: 3
		});
		expect(onsaved).toHaveBeenCalledWith(bread);
	});

	it('BR-PRD-01 shows a taken SKU under the SKU field', async () => {
		serve({
			'GET /api/v1/categories': categories,
			'POST /api/v1/products': () =>
				Response.json(
					{
						code: 'VALIDATION_FAILED',
						message: 'Periksa kembali isian Anda.',
						errors: [{ field: 'sku', code: 'TAKEN', message: 'Sudah dipakai. Pilih yang lain.' }],
						meta: { request_id: 'r' }
					},
					{ status: 422 }
				)
		});
		await render(ProductsHarness, { view: 'form' });

		await userEvent.type(page.getByLabelText('Nama produk'), 'Roti Sobek');
		await userEvent.selectOptions(page.getByLabelText('Kategori'), breadCategory);
		await userEvent.type(page.getByLabelText('SKU'), 'ROT-0001');
		await userEvent.click(page.getByRole('button', { name: 'Simpan produk' }));

		await expect.element(page.getByText('Sudah dipakai. Pilih yang lain.')).toBeVisible();
		await expect.element(page.getByLabelText('SKU')).toHaveFocus();
	});
});

describe('Product detail', () => {
	it('FR-PRD-03 archives and reactivates a product', async () => {
		serve({
			'GET /api/v1/categories': categories,
			[`GET /api/v1/products/${productId}`]: () => envelope(bread),
			[`POST /api/v1/products/${productId}/archive`]: () =>
				envelope({ ...bread, status: 'ARCHIVED' }),
			[`POST /api/v1/products/${productId}/reactivate`]: () => envelope(bread)
		});
		const { container } = await render(ProductsHarness, { view: 'detail', productId });

		await userEvent.click(page.getByRole('button', { name: 'Arsipkan' }));
		await expect.element(page.getByText('Produk ini diarsipkan.')).toBeVisible();
		await userEvent.click(page.getByRole('button', { name: 'Aktifkan kembali' }));

		await expect.element(page.getByText('Produk aktif kembali.')).toBeVisible();
		await expectAccessible(container);
	});
});
