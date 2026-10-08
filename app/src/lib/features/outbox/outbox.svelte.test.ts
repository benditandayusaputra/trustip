import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { outboxStore } from '#lib/offline/store.js';
import OutboxHarness from '#lib/testing/OutboxHarness.svelte';
import ReviewHarness from '#lib/testing/ReviewHarness.svelte';
import type { OutboxEntry } from '#lib/offline/outbox.js';
import { expectAccessible } from '#lib/testing/accessibility.js';
import { aQueuedDrop, emptyOutbox, storeCashier } from '#lib/testing/outbox.js';

function envelope(data: unknown, status = 200): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } }, { status });
}

function serve(routes: Record<string, () => Response>) {
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const respond = routes[`${init.method ?? 'GET'} ${path}`];
		return Promise.resolve(
			respond?.() ??
				Response.json(
					{ code: 'NOT_FOUND', message: 'x', meta: { request_id: 'r' } },
					{ status: 404 }
				)
		);
	});
}

const me = { 'GET /api/v1/me': () => envelope(storeCashier) };

afterEach(async () => {
	await emptyOutbox();
	vi.unstubAllGlobals();
});

describe('Outbox page (FR-PWA-09)', () => {
	it('lists the waiting entries with their status and counts them in the header', async () => {
		serve(me);
		await outboxStore.put(aQueuedDrop());
		await outboxStore.put(
			aQueuedDrop({
				id: 'entry-review',
				status: 'NEEDS_REVIEW',
				problem: {
					status: 409,
					code: 'STOCK_CHANGED',
					message: 'Stok sudah berubah.',
					requestId: 'r',
					fields: []
				}
			})
		);

		const { container } = await render(OutboxHarness);

		await expect.element(page.getByText('Menunggu sinyal')).toBeVisible();
		await expect.element(page.getByText('Galat terakhir: Stok sudah berubah.')).toBeVisible();
		await expect.element(page.getByRole('link', { name: 'Tinjau' })).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Antrean, 2 catatan menunggu' }))
			.toBeVisible();
		await expectAccessible(container);
	});

	it('sends a waiting entry on request and links the new document', async () => {
		serve({
			...me,
			'POST /api/v1/drops': () => envelope({ id: 'drop-9', number: 'TM-202611-0009' }, 201)
		});
		await outboxStore.put(aQueuedDrop());
		await render(OutboxHarness);

		await userEvent.click(page.getByRole('button', { name: 'Kirim sekarang' }));

		await expect.element(page.getByText('1 catatan terkirim.')).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Buka TM-202611-0009' }))
			.toHaveAttribute('href', '/app/titipan/drop-9');
	});

	it('discards an entry after asking', async () => {
		serve(me);
		await outboxStore.put(aQueuedDrop());
		await render(OutboxHarness);

		await userEvent.click(page.getByRole('button', { name: 'Buang' }));
		await userEvent.click(page.getByRole('dialog').getByRole('button', { name: 'Buang' }));

		await expect.element(page.getByText('Tidak ada catatan yang menunggu.')).toBeVisible();
		await expect(outboxStore.listFor(storeCashier.id)).resolves.toEqual([]);
	});
});

describe('Review screen (R-OFF-09)', () => {
	it('shows what the user entered beside the stock recorded now and links the fix', async () => {
		const partnershipId = '0192f0c4-0000-7000-8000-000000000501';
		serve({
			...me,
			[`GET /api/v1/partnerships/${partnershipId}/count-sheet`]: () =>
				envelope({
					partnership_id: partnershipId,
					is_active: true,
					agreement_version: 1,
					store: { id: 's1', name: 'Warung Bu Sri' },
					consignor: { id: 'c1', name: 'Roti Pak Anton' },
					last_document_at: null,
					products: [
						{
							product_id: 'bread',
							product_name: 'Roti Manis',
							sku: 'ROT-0001',
							unit: 'pcs',
							recorded_qty: 20,
							batches: []
						}
					]
				})
		});
		const refused = {
			...aQueuedDrop({ status: 'NEEDS_REVIEW' }),
			id: 'entry-count',
			kind: 'COUNT',
			problem: {
				status: 409,
				code: 'STOCK_CHANGED',
				message: 'Stok sudah berubah.',
				requestId: 'r',
				fields: []
			},
			request: {
				partnership_id: partnershipId,
				agreement_version: 1,
				counted_at: null,
				occasion: 'STORE_CLOSING',
				note: '',
				lines: [
					{
						product_id: 'bread',
						recorded_qty: 18,
						counted_qty: 2,
						returned_qty: 0,
						return_reason: null
					}
				]
			}
		} as OutboxEntry;
		await outboxStore.put(refused);

		const { container } = await render(ReviewHarness, { entryId: refused.id });

		await expect.element(page.getByText('Galat terakhir: Stok sudah berubah.')).toBeVisible();
		await expect.element(page.getByText('Stok tercatat sekarang').last()).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Perbaiki dan kirim ulang' }))
			.toHaveAttribute('href', `/app/hitung/baru?mitra=${partnershipId}&antrean=entry-count`);
		await expectAccessible(container);
	});
});
