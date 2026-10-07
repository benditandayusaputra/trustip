import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import CaptchaHarness from '#lib/testing/CaptchaHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';
import type { TurnstileApi, TurnstileRenderOptions } from './turnstile';

interface FakeTurnstile {
	api: TurnstileApi;
	options: () => TurnstileRenderOptions;
}

function installFakeTurnstile(): FakeTurnstile {
	const rendered: TurnstileRenderOptions[] = [];
	const api: TurnstileApi = {
		render: (_container, options) => {
			rendered.push(options);
			return 'widget-1';
		},
		reset: vi.fn(),
		remove: vi.fn()
	};
	window.turnstile = api;
	return {
		api,
		options: () => {
			const [first] = rendered;
			if (!first) {
				throw new Error('widget was not rendered');
			}
			return first;
		}
	};
}

afterEach(() => {
	delete window.turnstile;
	vi.restoreAllMocks();
});

describe('CaptchaField', () => {
	it('renders the widget with the site key, the action, and the dark theme', async () => {
		const fake = installFakeTurnstile();

		const { container } = await render(CaptchaHarness);

		await vi.waitFor(() => {
			expect(fake.options()).toMatchObject({
				sitekey: 'site-key',
				action: 'login',
				theme: 'dark',
				language: 'id'
			});
		});
		await expect.element(page.getByRole('group', { name: 'Verifikasi keamanan' })).toBeVisible();
		await expectAccessible(container);
	});

	it('hands over the solved token and drops it when it expires', async () => {
		const fake = installFakeTurnstile();
		await render(CaptchaHarness);
		await vi.waitFor(() => fake.options());

		fake.options().callback('solved-token');
		await expect.element(page.getByTestId('token')).toHaveTextContent('solved-token');
		fake.options()['expired-callback']();

		await expect.element(page.getByTestId('token')).toHaveTextContent('');
	});

	it('asks for a fresh challenge on reset', async () => {
		const fake = installFakeTurnstile();
		await render(CaptchaHarness);
		await vi.waitFor(() => fake.options());
		fake.options().callback('used-token');

		await userEvent.click(page.getByRole('button', { name: 'Ulangi' }));

		expect(fake.api.reset).toHaveBeenCalledWith('widget-1');
		await expect.element(page.getByTestId('token')).toHaveTextContent('');
	});

	it('says so when the widget cannot be loaded', async () => {
		vi.spyOn(document.head, 'append').mockImplementation((...nodes) => {
			nodes.forEach((node) => {
				if (node instanceof HTMLScriptElement) {
					node.dispatchEvent(new Event('error'));
				}
			});
		});

		await render(CaptchaHarness);

		await expect.element(page.getByRole('alert')).toHaveTextContent(/tidak bisa dimuat/u);
	});
});
