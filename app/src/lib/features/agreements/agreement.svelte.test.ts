import { page } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import AgreementHarness from '#lib/testing/AgreementHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const partnershipId = '0192f0c4-0000-7000-8000-000000000501';

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
			product_id: '0192f0c4-0000-7000-8000-000000000101',
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
	seal: {
		is_valid: true,
		events: [
			{
				event_type: 'AGREEMENT_ACCEPTED',
				actor_org_name: 'Warung Bu Sri',
				actor_user_name: 'Sri',
				occurred_at: '2026-10-19T09:00:00Z',
				code: 'V7KA-DYAY-2Y4Z-3FGX'
			}
		]
	}
};

function serve(data: unknown) {
	vi.stubGlobal('fetch', () =>
		Promise.resolve(Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } }))
	);
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Agreement page', () => {
	it('FR-SIG-02 shows a valid seal with its trail and verification code', async () => {
		serve(agreement);
		const { container } = await render(AgreementHarness, { partnershipId });

		await expect.element(page.getByText('Segel sah')).toBeVisible();
		await expect.element(page.getByText(/V7KA-DYAY-2Y4Z-3FGX/u)).toBeVisible();
		await expect.element(page.getByText(/Disetujui Sri \(Warung Bu Sri\)/u)).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-AGR-01 lists the three numbers per product', async () => {
		serve(agreement);
		await render(AgreementHarness, { partnershipId });

		await expect.element(page.getByText('Rp1.750').first()).toBeInTheDocument();
		await expect.element(page.getByText('Rp250').first()).toBeInTheDocument();
	});

	it('R-SEC-33 warns both parties when the seal no longer matches', async () => {
		serve({ ...agreement, seal: { ...agreement.seal, is_valid: false } });
		await render(AgreementHarness, { partnershipId });

		await expect.element(page.getByText('Segel tidak sah')).toBeVisible();
	});
});
