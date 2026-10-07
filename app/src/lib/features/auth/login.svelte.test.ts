import { page, userEvent } from 'vitest/browser';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import LoginHarness from '#lib/testing/LoginHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';
import { installFakeCaptcha, type FakeCaptcha } from '#lib/testing/captcha.js';

let captcha: FakeCaptcha;

function failWith(status: number, code: string, message: string) {
	return () =>
		Promise.resolve(Response.json({ code, message, meta: { request_id: 'r1' } }, { status }));
}

async function fillAndSubmit(): Promise<void> {
	await userEvent.fill(page.getByLabelText('Email'), 'sari@contoh.id');
	await userEvent.fill(page.getByLabelText('Kata sandi'), 'Kp7#mQz2vX');
	await vi.waitFor(() => {
		captcha.solve('solved-token');
	});
	await userEvent.click(page.getByRole('button', { name: 'Masuk' }));
}

beforeEach(() => {
	captcha = installFakeCaptcha();
});

afterEach(() => {
	delete window.turnstile;
	vi.unstubAllGlobals();
	vi.restoreAllMocks();
});

describe('LoginForm', () => {
	it('asks for every missing field and focuses the email', async () => {
		const { container } = await render(LoginHarness, { onsignedin: vi.fn() });

		await userEvent.click(page.getByRole('button', { name: 'Masuk' }));

		await expect.element(page.getByLabelText('Email')).toHaveFocus();
		await expect.element(page.getByText('Selesaikan verifikasi keamanan dulu.')).toBeVisible();
		await expectAccessible(container);
	});

	it('sends the credentials with the captcha token and hands over on success', async () => {
		const fetchMock = vi.fn(() =>
			Promise.resolve(Response.json({ code: 'OK', message: '', data: {}, meta: {} }))
		);
		vi.stubGlobal('fetch', fetchMock);
		const onsignedin = vi.fn();
		await render(LoginHarness, { onsignedin });

		await fillAndSubmit();

		await vi.waitFor(() => {
			expect(onsignedin).toHaveBeenCalledOnce();
		});
		const [path, init] = fetchMock.mock.calls[0] as unknown as [string, { body: string }];
		expect(path).toBe('/api/v1/sessions');
		expect(JSON.parse(init.body)).toEqual({
			email: 'sari@contoh.id',
			password: 'Kp7#mQz2vX',
			captcha_token: 'solved-token'
		});
	});

	it('R-SEC-08 shows one message for a wrong email or password and asks for a new captcha', async () => {
		vi.stubGlobal('fetch', failWith(401, 'LOGIN_FAILED', 'Email atau kata sandi salah.'));
		const onsignedin = vi.fn();
		await render(LoginHarness, { onsignedin });

		await fillAndSubmit();

		await expect.element(page.getByRole('alert')).toHaveTextContent('Email atau kata sandi salah.');
		await expect.element(page.getByRole('alert')).toHaveFocus();
		expect(captcha.resets()).toBe(1);
		expect(onsignedin).not.toHaveBeenCalled();
	});

	it('BR-AUTH-01 tells a locked account when it may try again', async () => {
		const locked =
			'Akun dikunci sementara karena terlalu banyak percobaan masuk. Coba lagi dalam 15 menit.';
		vi.stubGlobal('fetch', failWith(429, 'ACCOUNT_LOCKED', locked));
		await render(LoginHarness, { onsignedin: vi.fn() });

		await fillAndSubmit();

		await expect.element(page.getByRole('alert')).toHaveTextContent(locked);
	});
});
