import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import NeedsHarness from '#lib/testing/NeedsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const needId = '0192f0c4-0000-7000-8000-000000000001';
const snacksId = '01a1161e-0a49-71f7-a49a-3cf81042e3b0';

const need = {
	id: needId,
	store: { id: 's1', name: 'Kantin Kampus' },
	title: 'Butuh camilan kemasan kecil',
	category: { id: snacksId, name: 'Kue kering dan camilan' },
	description: 'Harga di bawah 5 ribu.',
	est_qty_per_week: 50,
	max_retail_price: 5000,
	preferred_model: 'COMMISSION',
	offered_commission_bps: 2000,
	status: 'OPEN',
	expires_at: '2026-11-19T03:00:00Z',
	created_at: '2026-10-20T03:00:00Z',
	response_count: 1
};

const detail = {
	...need,
	responses: [
		{
			proposal_id: '0192f0c4-0000-7000-8000-000000000301',
			number: 'AJ-202610-0001',
			status: 'OPEN',
			from_org_id: 'c1',
			awaiting_org_id: 's1',
			consignor_name: 'Keripik Bu Rina',
			created_at: '2026-10-21T03:00:00Z'
		}
	]
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
		sent[key] = typeof init.body === 'string' ? JSON.parse(init.body) : undefined;
		return Promise.resolve(
			routes[key]?.() ??
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

const categories = () => list([{ id: snacksId, name: 'Kue kering dan camilan', slug: 'kue' }]);

describe('Need form', () => {
	it('FR-NED-01 posts the need with its optional terms', async () => {
		const sent = serve({
			'GET /api/v1/categories': categories,
			'POST /api/v1/needs': () => envelope(need, 201)
		});
		const onposted = vi.fn();
		const { container } = await render(NeedsHarness, { view: 'form', onposted });

		await userEvent.type(
			page.getByRole('textbox', { name: 'Judul' }),
			'Butuh camilan kemasan kecil'
		);
		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Kategori' }), snacksId);
		await userEvent.type(page.getByRole('textbox', { name: 'Perkiraan jumlah per minggu' }), '50');
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Pasang Kebutuhan' }));

		await expect.element(page.getByText('Kebutuhan tayang di Papan Kebutuhan.')).toBeVisible();
		expect(sent['POST /api/v1/needs']).toMatchObject({
			title: 'Butuh camilan kemasan kecil',
			category_id: snacksId,
			est_qty_per_week: 50,
			preferred_model: null
		});
		expect(onposted).toHaveBeenCalled();
	});

	it('BR-NED-02 shows the server refusal of a phone number under the description', async () => {
		serve({
			'GET /api/v1/categories': categories,
			'POST /api/v1/needs': () =>
				Response.json(
					{
						code: 'VALIDATION_FAILED',
						message: 'Periksa kembali isian Anda.',
						errors: [
							{
								field: 'description',
								code: 'CONTACT_NOT_ALLOWED',
								message: 'Nomor telepon dan tautan tidak boleh ada di teks.'
							}
						],
						meta: { request_id: 'r' }
					},
					{ status: 422 }
				)
		});
		await render(NeedsHarness, { view: 'form' });

		await userEvent.type(
			page.getByRole('textbox', { name: 'Judul' }),
			'Butuh camilan kemasan kecil'
		);
		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Kategori' }), snacksId);
		await userEvent.click(page.getByRole('button', { name: 'Pasang Kebutuhan' }));

		await expect
			.element(page.getByText('Nomor telepon dan tautan tidak boleh ada di teks.'))
			.toBeVisible();
	});
});

describe('Need list and detail', () => {
	it('FR-NED-06 lists the needs with their responses count', async () => {
		serve({ 'GET /api/v1/needs': () => list([need]) });
		const { container } = await render(NeedsHarness, { view: 'list' });

		await expect.element(page.getByText('Butuh camilan kemasan kecil')).toBeVisible();
		await expect.element(page.getByText(/1 tanggapan/u)).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-NED-06 shows the responses and closes the need after confirming', async () => {
		const sent = serve({
			[`GET /api/v1/needs/${needId}`]: () => envelope(detail),
			[`POST /api/v1/needs/${needId}/close`]: () => envelope({ ...need, status: 'CLOSED' })
		});
		await render(NeedsHarness, { view: 'detail', needId });

		await expect.element(page.getByText('Keripik Bu Rina')).toBeVisible();
		await userEvent.click(page.getByRole('button', { name: 'Tutup Kebutuhan' }));
		await userEvent.click(
			page.getByRole('dialog').getByRole('button', { name: 'Tutup Kebutuhan' })
		);

		await expect.element(page.getByText('Kebutuhan ditutup.')).toBeVisible();
		await expect.element(page.getByText('Ditutup')).toBeVisible();
		expect(Object.keys(sent)).toContain(`POST /api/v1/needs/${needId}/close`);
	});
});
