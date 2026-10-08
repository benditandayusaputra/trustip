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
		rows: [{ key: 'a', label: 'Roti Pak Anton', ...figures }]
	};
}

function serve() {
	const requested: string[] = [];
	vi.stubGlobal('fetch', (path: string) => {
		requested.push(path);
		const groupBy = new URL(path, 'https://trustip.test').searchParams.get('group_by') ?? 'PARTNER';
		return Promise.resolve(
			Response.json({ code: 'OK', message: '', data: report(groupBy), meta: { request_id: 'r' } })
		);
	});
	return requested;
}

afterEach(() => {
	vi.useRealTimers();
	vi.unstubAllGlobals();
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
});
