import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ProposalsHarness from '#lib/testing/ProposalsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const partnershipId = '0192f0c4-0000-7000-8000-000000000501';
const breadId = '0192f0c4-0000-7000-8000-000000000101';
const chipsId = '0192f0c4-0000-7000-8000-000000000102';

const agreement = {
	id: '0192f0c4-0000-7000-8000-000000000401',
	partnership_id: partnershipId,
	version: 1,
	status: 'ACTIVE',
	number: 'AJ-202610-0001/V1',
	consignor: { id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c', name: 'Roti Pak Anton' },
	store: { id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d', name: 'Warung Bu Sri' },
	terms: { settlement_schedule: 'WEEKLY', payment_terms_days: 7, count_rhythm: 'DAILY', notes: '' },
	items: [
		{
			product_id: breadId,
			product_name: 'Roti Manis',
			sku: 'ROT-0001',
			unit: 'pcs',
			pricing_model: 'FIXED_DEPOSIT',
			retail_price: 2000,
			commission_bps: 0,
			deposit_price: 1750,
			is_active: true,
			per_unit: { gross: 2000, store_share: 250, consignor_share: 1750 }
		}
	],
	trail: {
		proposed_by: { name: 'Anton', organization_name: 'Roti Pak Anton' },
		accepted_by: { name: 'Sri', organization_name: 'Warung Bu Sri' },
		accepted_at: '2026-10-19T09:00:00Z',
		source_proposal_id: '0192f0c4-0000-7000-8000-000000000301',
		source_proposal_number: 'AJ-202610-0001'
	},
	seal: { is_valid: true, events: [] }
};

const chips = {
	product_id: chipsId,
	name: 'Keripik Pedas',
	sku: 'KER-0001',
	unit: 'bungkus',
	description: '',
	net_content: '',
	category_name: 'Keripik dan kerupuk',
	photo_url: null,
	is_active: false,
	terms: null
};

function ok(data: unknown, status = 200): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } }, { status });
}

function serve(extra: Record<string, () => Response> = {}) {
	const sent: Record<string, unknown> = {};
	const routes: Record<string, () => Response> = {
		[`GET /api/v1/partnerships/${partnershipId}/agreement`]: () => ok(agreement),
		[`GET /api/v1/partnerships/${partnershipId}/catalog`]: () =>
			Response.json({
				code: 'OK',
				message: '',
				data: [chips],
				meta: { request_id: 'r', next_cursor: null, has_more: false }
			}),
		[`POST /api/v1/partnerships/${partnershipId}/amendments`]: () => ok({ id: 'p2' }, 201),
		...extra
	};
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		sent[key] = typeof init.body === 'string' ? JSON.parse(init.body) : init.body;
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

describe('Amendment form', () => {
	it('FR-PRP-08 stops one product, adds another, and sends both changes', async () => {
		const sent = serve();
		const ondone = vi.fn();
		const { container } = await render(ProposalsHarness, {
			view: 'amend',
			proposalId: partnershipId,
			ondone
		});

		await userEvent.selectOptions(page.getByLabelText('Perubahan untuk Roti Manis'), 'STOP');
		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Pilih produk' }), chipsId);
		await userEvent.click(page.getByRole('button', { name: 'Tambahkan' }));
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Kirim Ajuan perubahan' }));

		await vi.waitFor(() => {
			expect(ondone).toHaveBeenCalled();
		});
		expect(sent[`POST /api/v1/partnerships/${partnershipId}/amendments`]).toMatchObject({
			terms: { payment_terms_days: 7 },
			items: [
				{ product_id: breadId, action: 'STOP' },
				{ product_id: chipsId, action: 'ADD' }
			]
		});
	});

	it('FR-AGR-06 prefills the quick add with the suggested prices', async () => {
		serve({
			[`GET /api/v1/products/${chipsId}`]: () =>
				ok({ id: chipsId, suggested_retail_price: 15000, suggested_deposit_price: 12000 })
		});
		await render(ProposalsHarness, { view: 'amend', proposalId: partnershipId, code: chipsId });

		await expect.element(page.getByRole('group', { name: 'Keripik Pedas' })).toBeVisible();
		await expect.element(page.getByText('Rp150.000')).toBeVisible();
	});
});
