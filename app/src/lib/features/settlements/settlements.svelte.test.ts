import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import SettlementsHarness from '#lib/testing/SettlementsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const partnershipId = '0192f0c4-0000-7000-8000-000000000501';
const settlementId = '0192f0c4-0000-7000-8000-000000000701';
const proofId = '0192f0c4-0000-7000-8000-000000000601';
const tinyPng =
	'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==';
const store = { id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d', name: 'Warung Bu Sri' };
const consignor = { id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c', name: 'Roti Pak Anton' };

function line(name: string, qty: number, split: [number, number, number, number?]) {
	const commission = split[3] ?? 0;
	return {
		product_id: `0192f0c4-0000-7000-8000-00000000010${String(qty % 10)}`,
		product_name: name,
		sku: 'SKU-0001',
		unit: 'pcs',
		pricing_model: commission > 0 ? 'COMMISSION' : 'FIXED_DEPOSIT',
		retail_price: split[0] / qty,
		commission_bps: commission,
		deposit_price: commission > 0 ? 0 : split[2] / qty,
		qty,
		split: { gross: split[0], store_share: split[1], consignor_share: split[2] }
	};
}

const settlement = {
	id: settlementId,
	number: 'ST-202611-0001',
	partnership_id: partnershipId,
	status: 'ISSUED',
	schedule: 'WEEKLY',
	period: { start: '2026-11-08T17:00:00Z', end: '2026-11-15T17:00:00Z' },
	due_date: '2026-11-23',
	issued_at: '2026-11-15T17:10:00Z',
	paid_at: null,
	standing: { is_overdue: false, days_late: 0, age: 'NOT_DUE' },
	store,
	consignor,
	lines: [
		line('Keripik Pedas 100 g', 27, [405000, 81000, 324000, 2000]),
		line('Kue Bawang 250 g', 3, [67500, 11813, 55687, 1750]),
		line('Sambal Roa', 10, [350000, 70000, 280000]),
		line('Roti Manis', 18, [36000, 4500, 31500])
	],
	sold_qty: 58,
	split: { gross: 858500, store_share: 167313, consignor_share: 691187 },
	counts: [
		{
			id: '0192f0c4-0000-7000-8000-000000000901',
			number: 'HS-202611-0003',
			counted_at: '2026-11-14T12:00:00Z'
		}
	],
	payments: [] as unknown[],
	payee: {
		payment_info: 'Transfer sebelum jam 5 sore',
		payment_qr_file_id: null,
		methods: ['CASH', 'TRANSFER'],
		bank_name: 'BCA',
		bank_account_number: '1234567890',
		bank_account_holder: 'Anton'
	},
	seal: { is_valid: true, events: [] }
};

const pending = {
	id: '0192f0c4-0000-7000-8000-000000000801',
	amount: 691187,
	method: 'CASH',
	paid_on: '2026-11-20',
	proof_file_id: null,
	status: 'PENDING_CONFIRMATION',
	rejection_reason: null,
	recorded_by_name: 'Sri',
	recorded_at: '2026-11-20T03:00:00Z',
	review: null,
	seal: { is_valid: true, events: [] }
};

const summary = {
	id: settlementId,
	number: 'ST-202611-0001',
	partnership_id: partnershipId,
	status: 'ISSUED',
	schedule: 'WEEKLY',
	period: settlement.period,
	due_date: '2026-11-23',
	issued_at: settlement.issued_at,
	sold_qty: 58,
	consignor_share: 691187,
	counterpart: { ...store, type: 'STORE' },
	standing: { is_overdue: true, days_late: 3, age: 'LATE_1_7' }
};

function envelope(data: unknown, meta: Record<string, unknown> = {}): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r', ...meta } });
}

function serve(routes: Record<string, () => Response>, sent: unknown[] = []) {
	const requested: string[] = [];
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		requested.push(key);
		if (typeof init.body === 'string') {
			sent.push(JSON.parse(init.body));
		}
		return Promise.resolve(
			routes[key]?.() ??
				Response.json(
					{ code: 'NOT_FOUND', message: 'x', meta: { request_id: 'r' } },
					{ status: 404 }
				)
		);
	});
	return requested;
}

const list = (items: unknown[]) => () => envelope(items, { next_cursor: null, has_more: false });

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Settlement page', () => {
	it('FR-STL-04 shows the section 13 totals with one line per product and the source count', async () => {
		serve({ [`GET /api/v1/settlements/${settlementId}`]: () => envelope(settlement) });
		const { container } = await render(SettlementsHarness, {
			view: 'detail',
			targetId: settlementId
		});

		await expect.element(page.getByText('Rp858.500').first()).toBeVisible();
		await expect.element(page.getByText('Rp167.313').first()).toBeVisible();
		await expect.element(page.getByText('Rp691.187').first()).toBeVisible();
		await expect.element(page.getByRole('link', { name: 'HS-202611-0003' })).toBeVisible();
		await expect.element(page.getByText('Mingguan')).toBeVisible();
		await expectAccessible(container);
	});
});

describe('Settlement list', () => {
	it('FR-STL-09 filters by age and marks late settlements in red with the days late', async () => {
		const requested = serve({
			'GET /api/v1/settlements': list([summary]),
			'GET /api/v1/settlements?age=LATE_1_7': list([summary])
		});
		const { container } = await render(SettlementsHarness, { view: 'list' });

		await expect.element(page.getByText('Lewat 3 hari')).toBeVisible();
		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Umur' }), 'LATE_1_7');

		await vi.waitFor(() => {
			expect(requested).toContain('GET /api/v1/settlements?age=LATE_1_7');
		});
		await expectAccessible(container);
	});

	it('FR-RPT-06 follows the report filters and searches by partner or number', async () => {
		const requested = serve({ 'GET /api/v1/settlements?status=ISSUED': list([summary]) });
		await render(SettlementsHarness, {
			view: 'list',
			narrowing: { status: 'ISSUED', age: '', query: '0001' }
		});

		await expect.element(page.getByText('Menampilkan 1 Setoran')).toBeVisible();
		await expect.element(page.getByRole('combobox', { name: 'Umur' })).not.toBeInTheDocument();
		expect(requested).toContain('GET /api/v1/settlements?status=ISSUED');
	});

	it('explains that settlements come from counts when there are none', async () => {
		serve({ 'GET /api/v1/settlements': list([]) });
		await render(SettlementsHarness, { view: 'list' });

		await expect
			.element(
				page.getByText(
					'Belum ada Setoran. Setoran terbit sendiri dari penjualan yang sudah dihitung.'
				)
			)
			.toBeVisible();
	});
});

describe('Running sales', () => {
	it('FR-STL-10 shows sales that are not in a settlement yet', async () => {
		serve({
			[`GET /api/v1/partnerships/${partnershipId}/running-sales`]: () =>
				envelope({
					partnership_id: partnershipId,
					has_sales: true,
					since: '2026-11-16T05:00:00Z',
					sold_qty: 18,
					split: { gross: 36000, store_share: 4500, consignor_share: 31500 }
				})
		});
		await render(SettlementsHarness, { view: 'running', targetId: partnershipId });

		await expect.element(page.getByText('Rp31.500')).toBeVisible();
		await expect.element(page.getByText('18', { exact: true })).toBeVisible();
	});
});

describe('Payments', () => {
	it('FR-STL-05 lets the store owner record cash with a photo of the receipt', async () => {
		const sent: unknown[] = [];
		serve(
			{
				[`GET /api/v1/settlements/${settlementId}`]: () => envelope(settlement),
				'POST /api/v1/files': () =>
					envelope({ id: proofId, url: `/api/v1/files/${proofId}`, content_type: 'image/png' }),
				[`POST /api/v1/settlements/${settlementId}/payments`]: () =>
					envelope({ ...settlement, payments: [pending] })
			},
			sent
		);
		const { container } = await render(SettlementsHarness, {
			view: 'detail',
			targetId: settlementId
		});

		await expect.element(page.getByText(/Serahkan uang tunai langsung ke penitip/u)).toBeVisible();
		await expect.element(page.getByText('Transfer sebelum jam 5 sore')).toBeVisible();
		await expectAccessible(container);
		await userEvent.upload(
			page.getByLabelText('Pilih foto atau PDF'),
			new File([Uint8Array.from(atob(tinyPng), (char) => char.charCodeAt(0))], 'nota.png', {
				type: 'image/png'
			})
		);
		await expect.element(page.getByRole('img', { name: 'Foto bukti bayar' })).toBeVisible();
		await userEvent.click(page.getByRole('button', { name: 'Simpan pembayaran' }));

		await expect.element(page.getByText('Menunggu penitip').first()).toBeVisible();
		expect(sent[0]).toMatchObject({ amount: 691187, method: 'CASH', proof_file_id: proofId });
	});

	it('BR-STL-05 asks for proof before sending any payment', async () => {
		const sent: unknown[] = [];
		serve({ [`GET /api/v1/settlements/${settlementId}`]: () => envelope(settlement) }, sent);
		await render(SettlementsHarness, { view: 'detail', targetId: settlementId });

		await userEvent.click(page.getByRole('button', { name: 'Simpan pembayaran' }));

		await expect
			.element(page.getByText('Unggah bukti pembayaran berupa foto atau PDF.'))
			.toBeVisible();
		expect(sent).toHaveLength(0);
	});

	it('offers only the methods the consignor accepts and shows the bank account for a transfer', async () => {
		const sent: unknown[] = [];
		serve(
			{
				[`GET /api/v1/settlements/${settlementId}`]: () => envelope(settlement),
				'POST /api/v1/files': () =>
					envelope({
						id: proofId,
						url: `/api/v1/files/${proofId}`,
						content_type: 'application/pdf'
					}),
				[`POST /api/v1/settlements/${settlementId}/payments`]: () =>
					envelope({ ...settlement, payments: [pending] })
			},
			sent
		);
		await render(SettlementsHarness, { view: 'detail', targetId: settlementId });
		const method = page.getByRole('combobox', { name: 'Metode' });

		await expect.element(method.getByRole('option')).toHaveLength(2);
		await userEvent.selectOptions(method, 'TRANSFER');
		await expect.element(page.getByText('1234567890')).toBeVisible();
		await expect.element(page.getByText('Anton', { exact: true })).toBeVisible();
		await userEvent.upload(
			page.getByLabelText('Pilih foto atau PDF'),
			new File(['%PDF-1.4'], 'struk-transfer.pdf', { type: 'application/pdf' })
		);
		await expect.element(page.getByText('struk-transfer.pdf')).toBeVisible();
		await userEvent.click(page.getByRole('button', { name: 'Simpan pembayaran' }));

		await expect.element(page.getByText('Menunggu penitip').first()).toBeVisible();
		expect(sent[0]).toMatchObject({ method: 'TRANSFER', proof_file_id: proofId });
	});

	it('FR-STL-06 lets the consignor owner accept the pending payment', async () => {
		const requested = serve({
			[`GET /api/v1/settlements/${settlementId}`]: () =>
				envelope({ ...settlement, payments: [pending] }),
			[`GET /api/v1/partnerships/${partnershipId}`]: () =>
				envelope({ counterpart: { whatsapp: '6281234567890' } }),
			[`POST /api/v1/payments/${pending.id}/confirm`]: () =>
				envelope({
					...settlement,
					status: 'PAID',
					standing: { is_overdue: false, days_late: 0, age: 'PAID' },
					payments: [{ ...pending, status: 'CONFIRMED' }]
				})
		});
		await render(SettlementsHarness, {
			view: 'detail',
			targetId: settlementId,
			viewerOrganizationId: consignor.id
		});

		await expect
			.element(page.getByRole('link', { name: /Ingatkan lewat WhatsApp/u }))
			.toHaveAttribute('href', expect.stringContaining('https://wa.me/6281234567890?text=Halo'));
		await userEvent.click(page.getByRole('button', { name: 'Uang diterima' }));
		await userEvent.click(page.getByRole('dialog').getByRole('button', { name: 'Uang diterima' }));

		await expect.element(page.getByText('Pembayaran diterima. Setoran lunas.')).toBeVisible();
		expect(requested).toContain(`POST /api/v1/payments/${pending.id}/confirm`);
	});

	it('FR-STL-06 needs a reason to reject', async () => {
		const requested = serve({
			[`GET /api/v1/settlements/${settlementId}`]: () =>
				envelope({ ...settlement, payments: [pending] })
		});
		await render(SettlementsHarness, {
			view: 'detail',
			targetId: settlementId,
			viewerOrganizationId: consignor.id
		});

		await userEvent.click(page.getByRole('button', { name: 'Tolak' }));
		await userEvent.click(page.getByRole('dialog').getByRole('button', { name: 'Tolak' }));

		await expect.element(page.getByText('Tulis alasan penolakan.')).toBeVisible();
		expect(requested).not.toContain(`POST /api/v1/payments/${pending.id}/reject`);
	});
});
