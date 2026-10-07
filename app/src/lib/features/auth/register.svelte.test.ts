import { page, userEvent } from 'vitest/browser';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import RegisterHarness from '#lib/testing/RegisterHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';
import type { TurnstileApi, TurnstileRenderOptions } from '#lib/components/ui/turnstile.js';

let solveCaptcha: (token: string) => void = () => undefined;
let resetCount = 0;

function installCaptcha(): void {
	const api: TurnstileApi = {
		render: (_container, options: TurnstileRenderOptions) => {
			solveCaptcha = options.callback;
			return 'widget-1';
		},
		reset: () => {
			resetCount += 1;
		},
		remove: () => undefined
	};
	window.turnstile = api;
}

function jsonResponse(status: number, body: unknown): Response {
	return Response.json(body, { status });
}

async function fillValidForm(): Promise<void> {
	await userEvent.fill(page.getByLabelText('Nama lengkap'), 'Sari Wulandari');
	await userEvent.fill(page.getByLabelText('Email'), 'sari@contoh.id');
	await userEvent.fill(page.getByLabelText('Kata sandi'), 'Kp7#mQz2vX');
	await userEvent.fill(page.getByLabelText('Nomor WhatsApp'), '0812 3456 7890');
	await userEvent.click(page.getByRole('checkbox'));
	await vi.waitFor(() => {
		solveCaptcha('solved-token');
	});
}

beforeEach(() => {
	resetCount = 0;
	installCaptcha();
});

afterEach(() => {
	delete window.turnstile;
	vi.unstubAllGlobals();
	vi.restoreAllMocks();
});

describe('RegisterForm', () => {
	it('shows every missing field and focuses the first one', async () => {
		const { container } = await render(RegisterHarness, { onregistered: vi.fn() });

		await userEvent.click(page.getByRole('button', { name: 'Daftar' }));

		await expect.element(page.getByLabelText('Nama lengkap')).toHaveFocus();
		await expect
			.element(page.getByText('Setujui kebijakan privasi dan ketentuan untuk mendaftar.'))
			.toBeVisible();
		await expect.element(page.getByText('Selesaikan verifikasi keamanan dulu.')).toBeVisible();
		await expectAccessible(container);
	});

	it('sends the form with the captcha token and hands over on success', async () => {
		const fetchMock = vi.fn(() =>
			Promise.resolve(jsonResponse(201, { code: 'OK', message: '', data: {}, meta: {} }))
		);
		vi.stubGlobal('fetch', fetchMock);
		const onregistered = vi.fn();
		await render(RegisterHarness, { onregistered });
		await fillValidForm();

		await userEvent.click(page.getByRole('button', { name: 'Daftar' }));

		await vi.waitFor(() => {
			expect(onregistered).toHaveBeenCalledOnce();
		});
		const [path, init] = fetchMock.mock.calls[0] as unknown as [string, { body: string }];
		expect(path).toBe('/api/v1/users');
		expect(JSON.parse(init.body)).toEqual({
			full_name: 'Sari Wulandari',
			email: 'sari@contoh.id',
			password: 'Kp7#mQz2vX',
			whatsapp: '0812 3456 7890',
			privacy_accepted: true,
			captcha_token: 'solved-token'
		});
		await expect.element(page.getByText('Akun dibuat. Silakan masuk.')).toBeVisible();
	});

	it('puts the server message under the email field and asks for a new captcha', async () => {
		vi.stubGlobal('fetch', () =>
			Promise.resolve(
				jsonResponse(422, {
					code: 'VALIDATION_FAILED',
					message: 'Periksa kembali isian Anda.',
					errors: [
						{
							field: 'email',
							code: 'EMAIL_TAKEN',
							message: 'Email sudah terdaftar. Silakan masuk.'
						}
					],
					meta: { request_id: 'r1' }
				})
			)
		);
		await render(RegisterHarness, { onregistered: vi.fn() });
		await fillValidForm();

		await userEvent.click(page.getByRole('button', { name: 'Daftar' }));

		await expect.element(page.getByLabelText('Email')).toHaveFocus();
		await expect.element(page.getByText('Email sudah terdaftar. Silakan masuk.')).toBeVisible();
		expect(resetCount).toBe(1);
	});

	it('keeps the input and offers a retry when the network fails', async () => {
		vi.stubGlobal('fetch', () => Promise.reject(new TypeError('Failed to fetch')));
		await render(RegisterHarness, { onregistered: vi.fn() });
		await fillValidForm();

		await userEvent.click(page.getByRole('button', { name: 'Daftar' }));

		await expect.element(page.getByRole('button', { name: 'Coba lagi' })).toBeVisible();
		await expect.element(page.getByLabelText('Nama lengkap')).toHaveValue('Sari Wulandari');
	});

	it('disables sending while the device is offline', async () => {
		vi.spyOn(navigator, 'onLine', 'get').mockReturnValue(false);

		await render(RegisterHarness, { onregistered: vi.fn() });

		await expect.element(page.getByRole('button', { name: 'Daftar' })).toBeDisabled();
		await expect.element(page.getByRole('status')).toHaveTextContent(/butuh koneksi/u);
	});
});
