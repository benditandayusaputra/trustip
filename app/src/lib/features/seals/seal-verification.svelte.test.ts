import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import type { SealVerification } from '#lib/api/public.js';
import SealVerificationView from './components/SealVerificationView.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const verification: SealVerification = {
	code: 'V7KA-DYAY-2Y4Z-3FGX',
	status: 'VALID',
	document_type: 'DROP',
	document_number: 'TM-202610-0007',
	event_type: 'DROP_RECORDED',
	occurred_at: '2026-10-19T01:00:00Z',
	consignor_name: 'Roti Mak Ijah',
	store_name: 'Warung Bu Tini'
};

describe('SealVerificationView', () => {
	it('FR-SIG-03 BR-SIG-06 shows the document, event, parties, and a valid seal', async () => {
		const { container } = await render(SealVerificationView, {
			code: 'v7kadyay2y4z3fgx',
			verification
		});

		await expect.element(page.getByText('Segel sah')).toBeVisible();
		await expect.element(page.getByText('TM-202610-0007')).toBeVisible();
		await expect
			.element(page.getByText('Roti Mak Ijah (penitip) dan Warung Bu Tini (toko)'))
			.toBeVisible();
		await expect.element(page.getByRole('heading', { name: 'V7KA-DYAY-2Y4Z-3FGX' })).toBeVisible();
		await expectAccessible(container);
	});

	it('shows an invalid seal in red words', async () => {
		await render(SealVerificationView, {
			code: verification.code,
			verification: { ...verification, status: 'INVALID' }
		});

		await expect.element(page.getByText('Segel tidak sah')).toBeVisible();
	});

	it('FR-SIG-03 tells the visitor an unknown code was not found', async () => {
		const { container } = await render(SealVerificationView, {
			code: 'AAAA-AAAA-AAAA-AAAA',
			verification: null
		});

		await expect.element(page.getByText('Kode tidak ditemukan')).toBeVisible();
		await expectAccessible(container);
	});
});
