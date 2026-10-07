import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import NotificationsHarness from '#lib/testing/NotificationsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const unread = {
	id: '0192f0c4-0000-7000-8000-000000000101',
	type: 'DROP_RECORDED',
	title: 'Titipan masuk',
	body: 'Warung Bu Tini mencatat 20 unit titipan masuk (TM-202610-0007).',
	link: '/app/titipan/0192f0c4-0000-7000-8000-000000000999',
	is_read: false,
	created_at: '2026-10-19T08:00:00Z'
};

function list(data: unknown[], hasMore = false): Response {
	return Response.json({
		code: 'OK',
		message: '',
		data,
		meta: { request_id: 'r', next_cursor: hasMore ? 'next' : null, has_more: hasMore }
	});
}

function ok(data: unknown): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } });
}

function serve(routes: Record<string, () => Response>) {
	const calls: string[] = [];
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		calls.push(key);
		return Promise.resolve(
			routes[key]?.() ??
				Response.json(
					{ code: 'NOT_FOUND', message: 'x', meta: { request_id: 'r' } },
					{ status: 404 }
				)
		);
	});
	return calls;
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Notifications', () => {
	it('FR-NTF-01 shows the unread count on the bell and lists notifications', async () => {
		serve({
			'GET /api/v1/notifications/unread-count': () => ok({ count: 1 }),
			'GET /api/v1/notifications': () => list([unread])
		});
		const { container } = await render(NotificationsHarness);

		await expect
			.element(page.getByRole('link', { name: 'Pemberitahuan, 1 belum dibaca' }))
			.toBeVisible();
		await expect.element(page.getByText('Titipan masuk', { exact: true })).toBeVisible();
		await expect.element(page.getByText(/Belum dibaca/u)).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-NTF-05 links each notification to its document', async () => {
		serve({
			'GET /api/v1/notifications/unread-count': () => ok({ count: 1 }),
			'GET /api/v1/notifications': () => list([unread])
		});
		await render(NotificationsHarness);

		await expect
			.element(page.getByRole('link', { name: /Titipan masuk/u }))
			.toHaveAttribute('href', unread.link);
	});

	it('R-SEC-64 never turns a stored link into a script link', async () => {
		serve({
			'GET /api/v1/notifications/unread-count': () => ok({ count: 1 }),
			'GET /api/v1/notifications': () => list([{ ...unread, link: 'javascript:alert(1)' }])
		});
		await render(NotificationsHarness);

		await expect
			.element(page.getByRole('link', { name: /Titipan masuk/u }))
			.toHaveAttribute('href', '/app');
	});

	it('FR-NTF-04 marks every notification read', async () => {
		const calls = serve({
			'GET /api/v1/notifications/unread-count': () => ok({ count: 1 }),
			'GET /api/v1/notifications': () => list([unread]),
			'POST /api/v1/notifications/read-all': () => ok(null)
		});
		await render(NotificationsHarness);

		await userEvent.click(page.getByRole('button', { name: 'Tandai semua dibaca' }));

		await expect.element(page.getByText('Semua pemberitahuan ditandai dibaca.')).toBeVisible();
		await expect.element(page.getByText(/Belum dibaca/u)).not.toBeInTheDocument();
		expect(calls).toContain('POST /api/v1/notifications/read-all');
	});

	it('shows an empty state with a way back', async () => {
		serve({
			'GET /api/v1/notifications/unread-count': () => ok({ count: 0 }),
			'GET /api/v1/notifications': () => list([])
		});
		const { container } = await render(NotificationsHarness);

		await expect.element(page.getByText(/Belum ada pemberitahuan/u)).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Pemberitahuan', exact: true }))
			.toBeVisible();
		await expectAccessible(container);
	});

	it('offers a retry when the list cannot be loaded', async () => {
		serve({
			'GET /api/v1/notifications/unread-count': () => ok({ count: 0 }),
			'GET /api/v1/notifications': () =>
				Response.json(
					{ code: 'SERVICE_UNAVAILABLE', message: 'Sibuk.', meta: { request_id: 'r' } },
					{ status: 503 }
				)
		});
		await render(NotificationsHarness);

		await expect.element(page.getByText('Pemberitahuan tidak bisa dimuat.')).toBeVisible();
	});
});
