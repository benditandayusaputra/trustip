import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { outboxStore } from '#lib/offline/store.js';
import OutboxHarness from '#lib/testing/OutboxHarness.svelte';
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
