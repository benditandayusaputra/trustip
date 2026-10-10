import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import AccountHarness from '#lib/testing/AccountHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';
import { cacheStore, outboxStore } from '#lib/offline/store.js';
import { aQueuedDrop, emptyOutbox } from '#lib/testing/outbox.js';

type Route = Record<string, () => Response>;

const sari = {
	id: '0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e',
	full_name: 'Sari Wulandari',
	email: 'sari@contoh.id',
	membership: {
		organization_id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c',
		organization_type: 'STORE',
		role: 'OWNER'
	}
};

const newcomer = { ...sari, membership: null };

function ok(data: unknown): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r1' } });
}

function failure(status: number, code: string): Response {
	return Response.json({ code, message: 'Gagal.', meta: { request_id: 'r1' } }, { status });
}

function serve(routes: Route) {
	const calls: string[] = [];
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		calls.push(key);
		const respond = routes[key];
		return Promise.resolve(respond ? respond() : failure(404, 'NOT_FOUND'));
	});
	return calls;
}

afterEach(async () => {
	await emptyOutbox(sari.id);
	vi.unstubAllGlobals();
});

describe('AccountGate', () => {
	it('FR-AUTH-06 greets the signed-in user by name', async () => {
		serve({ 'GET /api/v1/me': () => ok(sari) });

		const { container } = await render(AccountHarness, { onsignedout: vi.fn() });

		await expect.element(page.getByRole('heading', { name: 'Halo, Sari Wulandari' })).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-AUTH-05 sends a user without an organization away instead of showing the app', async () => {
		serve({ 'GET /api/v1/me': () => ok(newcomer) });
		const onmisplaced = vi.fn();

		await render(AccountHarness, { onsignedout: vi.fn(), onmisplaced });

		await vi.waitFor(() => {
			expect(onmisplaced).toHaveBeenCalledOnce();
		});
		await expect
			.element(page.getByRole('heading', { name: 'Halo, Sari Wulandari' }))
			.not.toBeInTheDocument();
	});

	it('sends a member away from a page meant for newcomers', async () => {
		serve({ 'GET /api/v1/me': () => ok(sari) });
		const onmisplaced = vi.fn();

		await render(AccountHarness, { placement: 'newcomer', onsignedout: vi.fn(), onmisplaced });

		await vi.waitFor(() => {
			expect(onmisplaced).toHaveBeenCalledOnce();
		});
	});

	it('hands over to sign in when the session cannot be refreshed', async () => {
		const calls = serve({
			'GET /api/v1/me': () => failure(401, 'UNAUTHENTICATED'),
			'POST /api/v1/sessions/refresh': () => failure(401, 'SESSION_EXPIRED')
		});
		const onsignedout = vi.fn();

		await render(AccountHarness, { onsignedout });

		await vi.waitFor(() => {
			expect(onsignedout).toHaveBeenCalledOnce();
		});
		expect(calls).toEqual(['GET /api/v1/me', 'POST /api/v1/sessions/refresh']);
	});

	it('FR-PUB-07 lets a visitor without a session stay on a public area and offers to sign in', async () => {
		serve({
			'GET /api/v1/me': () => failure(401, 'UNAUTHENTICATED'),
			'POST /api/v1/sessions/refresh': () => failure(401, 'SESSION_EXPIRED')
		});
		const onsignedout = vi.fn();

		await render(AccountHarness, { placement: 'anyone', onsignedout });

		await expect.element(page.getByRole('link', { name: 'Masuk' })).toBeVisible();
		await expect.element(page.getByRole('heading', { level: 1 })).toBeVisible();
		expect(onsignedout).not.toHaveBeenCalled();
	});

	it('offers a retry when the account cannot be loaded', async () => {
		let attempts = 0;
		serve({
			'GET /api/v1/me': () => (++attempts === 1 ? failure(503, 'SERVICE_UNAVAILABLE') : ok(sari))
		});
		await render(AccountHarness, { onsignedout: vi.fn() });

		await userEvent.click(page.getByRole('button', { name: 'Coba lagi' }));

		await expect.element(page.getByRole('heading', { name: 'Halo, Sari Wulandari' })).toBeVisible();
	});
});

describe('AccountMenu', () => {
	it('FR-AUTH-03 logs out this device', async () => {
		const calls = serve({
			'GET /api/v1/me': () => ok(sari),
			'POST /api/v1/sessions/logout': () => ok(null)
		});
		const onsignedout = vi.fn();
		await render(AccountHarness, { onsignedout });

		await userEvent.click(page.getByText('Akun', { exact: true }));
		await userEvent.click(page.getByRole('button', { name: 'Keluar', exact: true }));

		await vi.waitFor(() => {
			expect(onsignedout).toHaveBeenCalledOnce();
		});
		expect(calls).toContain('POST /api/v1/sessions/logout');
	});

	it('BR-PWA-02 forgets the data saved on this device when logging out', async () => {
		serve({
			'GET /api/v1/me': () => ok(sari),
			'POST /api/v1/sessions/logout': () => ok(null)
		});
		const onsignedout = vi.fn();
		await render(AccountHarness, { onsignedout });
		await expect.element(page.getByRole('heading', { name: 'Halo, Sari Wulandari' })).toBeVisible();
		await vi.waitFor(async () => {
			await expect(cacheStore.get('/api/v1/me')).resolves.toBeDefined();
		});

		await userEvent.click(page.getByText('Akun', { exact: true }));
		await userEvent.click(page.getByRole('button', { name: 'Keluar', exact: true }));

		await vi.waitFor(() => {
			expect(onsignedout).toHaveBeenCalledOnce();
		});
		await expect(cacheStore.get('/api/v1/me')).resolves.toBeUndefined();
	});

	it('BR-PWA-08 asks what to do with unsent entries and discards them on request', async () => {
		serve({
			'GET /api/v1/me': () => ok(sari),
			'POST /api/v1/sessions/logout': () => ok(null)
		});
		await outboxStore.put(aQueuedDrop({ userId: sari.id }));
		const onsignedout = vi.fn();
		await render(AccountHarness, { onsignedout });

		await userEvent.click(page.getByText('Akun', { exact: true }));
		await userEvent.click(page.getByRole('button', { name: 'Keluar', exact: true }));
		const dialog = page.getByRole('dialog', { name: 'Masih ada 1 catatan belum terkirim' });
		await expect.element(dialog).toBeVisible();
		expect(onsignedout).not.toHaveBeenCalled();
		await userEvent.click(dialog.getByRole('button', { name: 'Buang catatan' }));

		await vi.waitFor(() => {
			expect(onsignedout).toHaveBeenCalledOnce();
		});
		await expect(outboxStore.listFor(sari.id)).resolves.toEqual([]);
	});

	it('R-OFF-16 keeps unsent entries on the device when asked', async () => {
		serve({
			'GET /api/v1/me': () => ok(sari),
			'POST /api/v1/sessions/logout': () => ok(null)
		});
		await outboxStore.put(aQueuedDrop({ userId: sari.id }));
		const onsignedout = vi.fn();
		await render(AccountHarness, { onsignedout });

		await userEvent.click(page.getByText('Akun', { exact: true }));
		await userEvent.click(page.getByRole('button', { name: 'Keluar', exact: true }));
		await userEvent.click(page.getByRole('button', { name: 'Simpan di perangkat' }));

		await vi.waitFor(() => {
			expect(onsignedout).toHaveBeenCalledOnce();
		});
		await expect(outboxStore.listFor(sari.id)).resolves.toHaveLength(1);
	});

	it('FR-AUTH-03 asks for confirmation before logging out every device', async () => {
		const calls = serve({
			'GET /api/v1/me': () => ok(sari),
			'POST /api/v1/sessions/logout-all': () => ok(null)
		});
		const onsignedout = vi.fn();
		await render(AccountHarness, { onsignedout });

		await userEvent.click(page.getByText('Akun', { exact: true }));
		await userEvent.click(page.getByRole('button', { name: 'Keluar dari semua perangkat' }));
		const dialog = page.getByRole('dialog', { name: 'Keluar dari semua perangkat?' });
		await expect.element(dialog).toBeVisible();
		expect(calls).not.toContain('POST /api/v1/sessions/logout-all');
		await userEvent.click(dialog.getByRole('button', { name: 'Keluar dari semua perangkat' }));

		await vi.waitFor(() => {
			expect(onsignedout).toHaveBeenCalledOnce();
		});
		expect(calls).toContain('POST /api/v1/sessions/logout-all');
	});
});
