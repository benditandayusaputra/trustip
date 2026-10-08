import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ReportHarness from '#lib/testing/ReportHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

interface Sent {
	path: string;
	body: unknown;
}

function serve(response: () => Response) {
	const sent: Sent[] = [];
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		sent.push({ path, body: typeof init.body === 'string' ? JSON.parse(init.body) : init.body });
		return Promise.resolve(response());
	});
	return sent;
}

const created = () =>
	Response.json(
		{
			code: 'OK',
			message: '',
			data: { id: '0192f0c4-0000-7000-8000-00000000aaaa', status: 'OPEN' },
			meta: { request_id: 'r' }
		},
		{ status: 201 }
	);

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Report listing', () => {
	it('FR-TRS-04 sends the chosen reason with a note', async () => {
		const sent = serve(created);
		const { container } = await render(ReportHarness, {});

		await userEvent.click(page.getByRole('button', { name: 'Laporkan Roti Sobek' }));
		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Alasan' }), 'MISLEADING');
		await userEvent.fill(
			page.getByRole('textbox', { name: 'Keterangan' }),
			'Fotonya bukan roti ini.'
		);
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Kirim laporan' }));

		await expect
			.element(page.getByText('Laporan terkirim. Tim kami akan meninjaunya.'))
			.toBeVisible();
		expect(sent[0]).toEqual({
			path: '/api/v1/listing-reports',
			body: {
				target_type: 'PRODUCT',
				target_id: '0192f0c4-0000-7000-8000-000000000101',
				reason: 'MISLEADING',
				note: 'Fotonya bukan roti ini.'
			}
		});
	});

	it('FR-TRS-04 asks for a reason and for a note when the reason is Lainnya', async () => {
		const sent = serve(created);
		await render(ReportHarness, {});

		await userEvent.click(page.getByRole('button', { name: 'Laporkan Roti Sobek' }));
		await userEvent.click(page.getByRole('button', { name: 'Kirim laporan' }));
		await expect.element(page.getByText('Pilih alasan laporan.')).toBeVisible();
		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Alasan' }), 'OTHER');
		await userEvent.click(page.getByRole('button', { name: 'Kirim laporan' }));

		await expect.element(page.getByText('Tulis keterangan untuk alasan Lainnya.')).toBeVisible();
		expect(sent).toHaveLength(0);
	});

	it('shows the server message when the report cannot be sent', async () => {
		serve(() =>
			Response.json(
				{
					code: 'RATE_LIMITED',
					message: 'Terlalu banyak percobaan. Coba lagi dalam 60 menit.',
					meta: { request_id: 'r' }
				},
				{ status: 429, headers: { 'retry-after': '3600' } }
			)
		);
		await render(ReportHarness, {});

		await userEvent.click(page.getByRole('button', { name: 'Laporkan Roti Sobek' }));
		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Alasan' }), 'SPAM');
		await userEvent.click(page.getByRole('button', { name: 'Kirim laporan' }));

		await expect
			.element(page.getByText('Terlalu banyak percobaan. Coba lagi dalam 60 menit.'))
			.toBeVisible();
	});
});
