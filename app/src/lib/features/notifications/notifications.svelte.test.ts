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

class FakeStream {
	static opened: FakeStream[] = [];
	onerror: (() => void) | null = null;
	readonly #listeners = new Map<string, (event: MessageEvent) => void>();

	constructor() {
		FakeStream.opened.push(this);
	}

	addEventListener(name: string, listener: (event: MessageEvent) => void): void {
		this.#listeners.set(name, listener);
	}

	emit(name: string, data: unknown): void {
		this.#listeners.get(name)?.(new MessageEvent(name, { data: JSON.stringify(data) }));
	}

	close(): void {
		this.#listeners.clear();
	}
}

afterEach(() => {
	vi.unstubAllGlobals();
	FakeStream.opened = [];
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
		await expect.element(page.getByText('1 belum dibaca')).toBeVisible();
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
		let unreadCount = 1;
		const calls = serve({
			'GET /api/v1/notifications/unread-count': () => ok({ count: unreadCount }),
			'GET /api/v1/notifications': () => list([unread]),
			'POST /api/v1/notifications/read-all': () => {
				unreadCount = 0;
				return ok(null);
			}
		});
		await render(NotificationsHarness);
		await expect
			.element(page.getByRole('link', { name: 'Pemberitahuan, 1 belum dibaca' }))
			.toBeVisible();

		await userEvent.click(page.getByRole('button', { name: 'Tandai semua telah dibaca' }));

		await expect.element(page.getByText('Semua pemberitahuan ditandai dibaca.')).toBeVisible();
		await expect.element(page.getByText('1 belum dibaca')).not.toBeInTheDocument();
		await expect
			.element(page.getByRole('button', { name: /Tandai dibaca/u }))
			.not.toBeInTheDocument();
		expect(calls).toContain('POST /api/v1/notifications/read-all');
		await expect
			.element(page.getByRole('link', { name: 'Pemberitahuan', exact: true }))
			.toBeVisible();
	});

	it('FR-NTF-04 marks one notification read and filters the unread ones', async () => {
		const read = {
			...unread,
			id: '0192f0c4-0000-7000-8000-000000000102',
			title: 'Ajuan kerja sama',
			is_read: true
		};
		const calls = serve({
			'GET /api/v1/notifications/unread-count': () => ok({ count: 1 }),
			'GET /api/v1/notifications': () => list([unread, read]),
			[`POST /api/v1/notifications/${unread.id}/read`]: () => ok(null)
		});
		const { container } = await render(NotificationsHarness);

		await userEvent.click(page.getByRole('button', { name: 'Belum dibaca' }));
		await expect.element(page.getByText('Ajuan kerja sama')).not.toBeInTheDocument();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Tandai dibaca: Titipan masuk' }));

		await expect.element(page.getByText('Semua pemberitahuan sudah dibaca.')).toBeVisible();
		expect(calls).toContain(`POST /api/v1/notifications/${unread.id}/read`);
		await userEvent.click(page.getByRole('button', { name: 'Semua' }));
		await expect.element(page.getByText('Ajuan kerja sama')).toBeVisible();
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

	it('FR-NTF-02 shows a live notification at once and opens its document', async () => {
		const arrived = {
			...unread,
			id: '0192f0c4-0000-7000-8000-000000000103',
			title: 'Ajuan kerja sama',
			body: 'Roti Mas Bendi mengajukan kerja sama.',
			link: '/app/ajuan/0192f0c4-0000-7000-8000-000000000555'
		};
		let items = [unread];
		const calls = serve({
			'GET /api/v1/notifications/unread-count': () =>
				ok({ count: items.filter((item) => !item.is_read).length }),
			'GET /api/v1/notifications': () => list(items),
			'POST /api/v1/stream-tickets': () => ok({ ticket: 't', expires_at: '2026-11-02T00:00:30Z' }),
			[`POST /api/v1/notifications/${arrived.id}/read`]: () => ok(null)
		});
		vi.stubGlobal('EventSource', FakeStream);
		await render(NotificationsHarness, { streamUrl: 'https://api.trustip.test/api/v1/stream' });
		await expect.poll(() => FakeStream.opened.length, { timeout: 5000 }).toBe(1);

		items = [arrived, unread];
		FakeStream.opened[0]?.emit('notification', { id: arrived.id, type: 'PROPOSAL_RECEIVED' });

		await expect.element(page.getByText(arrived.body).first()).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Pemberitahuan, 2 belum dibaca' }))
			.toBeVisible();
		await expect.element(page.getByText('2 belum dibaca', { exact: true })).toBeVisible();
		const open = page.getByRole('link', { name: 'Lihat' });
		await expect.element(open).toHaveAttribute('href', arrived.link);
		document.addEventListener(
			'click',
			(event) => {
				event.preventDefault();
			},
			{ capture: true, once: true }
		);
		await userEvent.click(open);

		await expect.element(open).not.toBeInTheDocument();
		await expect.poll(() => calls).toContain(`POST /api/v1/notifications/${arrived.id}/read`);
	});
});
