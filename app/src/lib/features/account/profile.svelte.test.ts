import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import type { AccountProfile } from '#lib/api/account.js';
import AccountCardHarness from '#lib/testing/AccountCardHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const avatarId = '0192f0c4-aa00-7e66-a1bb-4d2f3e5d6a92';
const tinyPng =
	'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==';

const sari: AccountProfile = {
	id: '0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e',
	full_name: 'Sari Wulandari',
	email: 'sari@contoh.id',
	whatsapp: '6281234567890',
	avatar: null,
	membership: {
		organization_id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c',
		organization_type: 'STORE',
		role: 'OWNER'
	}
};

function ok(data: unknown, status = 200): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } }, { status });
}

type Route = (body: unknown) => Response;

function serve(routes: Record<string, Route>) {
	const sent: Record<string, unknown> = {};
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		sent[key] = typeof init.body === 'string' ? JSON.parse(init.body) : init.body;
		const respond = routes[key];
		return Promise.resolve(
			respond
				? respond(sent[key])
				: Response.json(
						{ code: 'NOT_FOUND', message: 'x', meta: { request_id: 'r' } },
						{ status: 404 }
					)
		);
	});
	return sent;
}

function saved(body: unknown): Response {
	const change = body as { full_name: string; whatsapp: string; avatar_file_id: string | null };
	return ok({
		...sari,
		full_name: change.full_name,
		whatsapp: change.whatsapp.replace(/^0/u, '62').replaceAll(' ', ''),
		avatar: change.avatar_file_id
			? { id: change.avatar_file_id, url: `/api/v1/files/${avatarId}` }
			: null,
		membership: null
	});
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('AccountCard', () => {
	it('FR-AUTH-13 edits the name and WhatsApp number of the signed-in user', async () => {
		const sent = serve({ 'PATCH /api/v1/me': saved });
		const { container } = await render(AccountCardHarness, { account: sari });

		await expect.element(page.getByText('081234567890')).toBeVisible();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Edit akun' }));
		await userEvent.fill(page.getByLabelText('Nama lengkap'), 'Sari W');
		await userEvent.fill(page.getByLabelText('Nomor WhatsApp'), '0813 1111 2222');
		await userEvent.click(page.getByRole('button', { name: 'Simpan akun' }));

		await expect.element(page.getByText('Akun tersimpan.')).toBeVisible();
		await expect.element(page.getByRole('heading', { name: 'Sari W' })).toBeVisible();
		expect(sent['PATCH /api/v1/me']).toEqual({
			full_name: 'Sari W',
			whatsapp: '0813 1111 2222',
			avatar_file_id: null
		});
	});

	it('FR-AUTH-13 uploads a private avatar and keeps the other fields', async () => {
		const sent = serve({
			'POST /api/v1/files': () => ok({ id: avatarId, url: `/api/v1/files/${avatarId}` }, 201),
			'PATCH /api/v1/me': saved
		});
		await render(AccountCardHarness, { account: sari });

		const bytes = Uint8Array.from(atob(tinyPng), (char) => char.charCodeAt(0));
		await userEvent.upload(
			page.getByLabelText('Ubah foto profil'),
			new File([bytes], 'foto.png', { type: 'image/png' })
		);

		await expect
			.element(page.getByRole('img', { name: 'Foto profil Sari Wulandari' }))
			.toBeVisible();
		expect((sent['POST /api/v1/files'] as FormData).get('purpose')).toBe('AVATAR');
		expect(sent['PATCH /api/v1/me']).toEqual({
			full_name: 'Sari Wulandari',
			whatsapp: '6281234567890',
			avatar_file_id: avatarId
		});
	});

	it('FR-AUTH-14 shows a wrong current password and then changes it', async () => {
		let attempts = 0;
		serve({
			'POST /api/v1/me/password': () => {
				attempts++;
				return attempts === 1
					? Response.json(
							{
								code: 'VALIDATION_FAILED',
								message: 'Periksa kembali isian Anda.',
								errors: [
									{
										field: 'current_password',
										code: 'WRONG_PASSWORD',
										message: 'Kata sandi saat ini salah.'
									}
								],
								meta: { request_id: 'r' }
							},
							{ status: 422 }
						)
					: ok(null);
			}
		});
		const { container } = await render(AccountCardHarness, { account: sari });

		await userEvent.click(page.getByRole('button', { name: 'Ubah sandi' }));
		await userEvent.fill(page.getByLabelText('Kata sandi saat ini'), 'Bukan-sandinya-1');
		await userEvent.fill(page.getByLabelText('Kata sandi baru'), 'Baru-sandi-2026');
		await userEvent.click(page.getByRole('button', { name: 'Simpan kata sandi' }));
		await expect.element(page.getByText('Kata sandi saat ini salah.')).toBeVisible();
		await expectAccessible(container);
		await userEvent.fill(page.getByLabelText('Kata sandi saat ini'), 'Kp7#mQz2vX');
		await userEvent.click(page.getByRole('button', { name: 'Simpan kata sandi' }));

		await expect
			.element(page.getByText('Kata sandi diganti. Sesi di perangkat lain sudah diakhiri.'))
			.toBeVisible();
		await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();
	});
});
