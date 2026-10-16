import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import NeedBoardHarness from '#lib/testing/NeedBoardHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const boardPath = '/api/v1/directory/needs';
const storeId = '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6b01';

const need = {
	id: '0192f0c4-0000-7000-8000-000000000901',
	title: 'Butuh camilan kemasan kecil',
	description: 'Untuk kantin kampus.',
	category: { id: 'c1', name: 'Kue kering dan camilan' },
	est_qty_per_week: 50,
	max_retail_price: 5000,
	preferred_model: 'COMMISSION',
	offered_commission_bps: 2000,
	expires_at: '2026-11-19T03:00:00Z',
	created_at: '2026-10-20T03:00:00Z',
	store: {
		id: storeId,
		slug: 'kantin',
		name: 'Kantin Kampus',
		city_name: 'Kota Semarang',
		store_type: 'CANTEEN'
	},
	distance_m: 1200,
	is_own: false
};

function list(data: unknown[]): Response {
	return Response.json({
		code: 'OK',
		message: '',
		data,
		meta: {
			request_id: 'r',
			next_cursor: null,
			has_more: false,
			search: {
				origin: { kind: 'ORGANIZATION', is_known: true },
				sort: 'NEAREST',
				selected_radius_km: 10,
				effective_radius_km: 10,
				is_expanded: false,
				category_ids: []
			}
		}
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

describe('Need board', () => {
	it('FR-NED-03 FR-NED-05 shows nearby open needs with a link to answer by proposal', async () => {
		serve({ [boardPath]: () => list([need]) });
		const { container } = await render(NeedBoardHarness, {});

		await expect
			.element(page.getByRole('heading', { name: 'Butuh camilan kemasan kecil' }))
			.toBeVisible();
		await expect.element(page.getByText('Harga jual sampai Rp5.000')).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Tanggapi dengan Ajuan' }))
			.toHaveAttribute('href', `/app/ajuan/baru?sasaran=${storeId}&kebutuhan=${need.id}`);
		await expectAccessible(container);
	});

	it('hides the answer link for stores and for the own need', async () => {
		serve({ [boardPath]: () => list([need]) });
		await render(NeedBoardHarness, { canRespond: false });

		await expect
			.element(page.getByRole('heading', { name: 'Butuh camilan kemasan kecil', exact: true }))
			.toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Tanggapi dengan Ajuan' }))
			.not.toBeInTheDocument();
	});

	it('FR-NED-03 FR-DIR-12 sorting by newest asks again and updates the address', async () => {
		const requested = serve({ [boardPath]: () => list([need]) });
		const addresses: string[] = [];
		await render(NeedBoardHarness, { onaddress: (address: string) => addresses.push(address) });

		await userEvent.click(page.getByRole('button', { name: 'Filter dan urutan' }));
		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Urutkan' }), 'NEWEST');

		await vi.waitFor(() => {
			expect(requested).toContain(`${boardPath}?sort=NEWEST`);
		});
		expect(addresses).toContain('sort=NEWEST');
	});
});
