import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ReportsHarness from '#lib/testing/ReportsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const figures = {
	sold_qty: 18,
	returned_qty: 2,
	received_qty: 20,
	returned_bps: 1000,
	split: { gross: 36000, store_share: 4500, consignor_share: 31500 }
};

function report(groupBy: string) {
	return {
		from: '2026-11-14',
		to: '2026-11-20',
		group_by: groupBy,
		summary: figures,
		rows: [
			groupBy === 'DAY'
				? { key: '2026-11-20', label: '2026-11-20', ...figures }
				: { key: 'a', label: 'Roti Pak Anton', ...figures }
		]
	};
}

function recapRow(name: string, paid: number) {
	return { partner_id: name, name, ...figures, paid };
}

const tally = (count: number, amount: number) => ({ count, amount });

function dataFor(path: string): unknown {
	const url = new URL(path, 'https://trustip.test');
	if (url.pathname.endsWith('/daily')) {
		const rows = [
			recapRow('Roti Pak Anton', 31500),
			recapRow('Keripik Bu Ani', 0),
			recapRow('Sambal Roa', 0)
		];
		return { date: '2026-11-20', total: recapRow('', 31500), rows };
	}
	if (url.pathname.endsWith('/settlements')) {
		const row = {
			partner_id: 'a',
			name: 'Warung Bu Sri',
			issued: tally(2, 45500),
			paid: tally(1, 31500),
			unpaid: tally(1, 14000),
			overdue: tally(1, 14000)
		};
		return { from: '2026-11-01', to: '2026-11-20', total: row, rows: [row] };
	}
	return report(url.searchParams.get('group_by') ?? 'PARTNER');
}

function serve() {
	const requested: string[] = [];
	vi.stubGlobal('fetch', (path: string) => {
		requested.push(path);
		return Promise.resolve(
			Response.json({ code: 'OK', message: '', data: dataFor(path), meta: { request_id: 'r' } })
		);
	});
	return requested;
}

afterEach(() => {
	vi.useRealTimers();
	vi.unstubAllGlobals();
	vi.restoreAllMocks();
});

describe('Sales report', () => {
	it('FR-RPT-04 shows the summary and rows for the last seven days', async () => {
		vi.useFakeTimers({ now: new Date('2026-11-20T05:00:00Z'), toFake: ['Date'] });
		const requested = serve();
		const { container } = await render(ReportsHarness, {});

		await expect.element(page.getByText('Rp36.000').first()).toBeVisible();
		await expect.element(page.getByText('Harus setor').first()).toBeVisible();
		await expect.element(page.getByText('10%').first()).toBeVisible();
		expect(requested[0]).toBe(
			'/api/v1/reports/sales?from=2026-11-14&to=2026-11-20&group_by=PARTNER'
		);
		await expectAccessible(container);
	});

	it('FR-RPT-07 switches to last month and to grouping per product', async () => {
		vi.useFakeTimers({ now: new Date('2026-11-20T05:00:00Z'), toFake: ['Date'] });
		const requested = serve();
		await render(ReportsHarness, { organizationType: 'CONSIGNOR' });

		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Rentang' }), 'LAST_MONTH');
		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Kelompokkan' }), 'PRODUCT');

		await vi.waitFor(() => {
			expect(requested).toContain(
				'/api/v1/reports/sales?from=2026-10-01&to=2026-10-31&group_by=PRODUCT'
			);
		});
		await expect.element(page.getByText('Hak penitip').first()).toBeVisible();
	});

	it('FR-RPT-01 shows the daily recap with one row per consignor', async () => {
		vi.useFakeTimers({ now: new Date('2026-11-20T05:00:00Z'), toFake: ['Date'] });
		const requested = serve();
		const { container } = await render(ReportsHarness, {});

		await userEvent.click(page.getByRole('tab', { name: 'Rekap harian' }));

		await expect.element(page.getByText('Keripik Bu Ani').last()).toBeVisible();
		await expect.element(page.getByText('Sudah dibayar').first()).toBeVisible();
		expect(requested).toContain('/api/v1/reports/daily?date=2026-11-20');
		await expectAccessible(container);
	});

	it('FR-RPT-06 shows issued, paid, unpaid and overdue settlements', async () => {
		vi.useFakeTimers({ now: new Date('2026-11-20T05:00:00Z'), toFake: ['Date'] });
		const requested = serve();
		await render(ReportsHarness, { organizationType: 'CONSIGNOR' });

		await expect.element(page.getByRole('tab', { name: 'Rekap harian' })).not.toBeInTheDocument();
		await userEvent.click(page.getByRole('tab', { name: 'Setoran' }));

		await expect.element(page.getByText('2 · Rp45.500').first()).toBeVisible();
		await expect.element(page.getByText('Lewat tempo').first()).toBeVisible();
		expect(requested).toContain('/api/v1/reports/settlements?from=2026-11-01&to=2026-11-20');
	});

	it('FR-RPT-05 draws the daily chart and offers it as a table', async () => {
		vi.useFakeTimers({ now: new Date('2026-11-20T05:00:00Z'), toFake: ['Date'] });
		const requested = serve();
		const { container } = await render(ReportsHarness, {});

		await expect.element(page.getByRole('img', { name: /Omzet per hari/u })).toBeVisible();
		expect(requested).toContain('/api/v1/reports/sales?from=2026-11-14&to=2026-11-20&group_by=DAY');
		await userEvent.click(page.getByRole('button', { name: 'Lihat sebagai tabel' }));

		await expect.element(page.getByRole('button', { name: 'Lihat sebagai grafik' })).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-RPT-08 exports the rows on screen as CSV', async () => {
		vi.useFakeTimers({ now: new Date('2026-11-20T05:00:00Z'), toFake: ['Date'] });
		serve();
		const blobs: Blob[] = [];
		vi.spyOn(URL, 'createObjectURL').mockImplementation((blob) => {
			blobs.push(blob as Blob);
			return 'blob:trustip';
		});
		vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined);
		const click = vi
			.spyOn(HTMLAnchorElement.prototype, 'click')
			.mockImplementation(() => undefined);
		await render(ReportsHarness, {});

		await userEvent.click(page.getByRole('button', { name: 'Unduh CSV' }));

		const csv = await blobs[0]?.text();
		expect(click).toHaveBeenCalled();
		expect(csv).toContain(
			'Kelompok,Unit terjual,Omzet,Untung toko,Harus setor,Unit dikembalikan,Persen dikembalikan'
		);
		expect(csv).toContain('Roti Pak Anton,18,36000,4500,31500,2,10%');
	});
});
