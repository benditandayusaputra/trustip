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
			const first = rendered.at(-1);
			if (!first) {
				throw new Error('widget was not rendered');
			}
			return first;
		}
	};
}

afterEach(() => {
	delete window.turnstile;
	delete document.documentElement.dataset.appearance;
	vi.restoreAllMocks();
});

describe('CaptchaField', () => {
	it('renders the widget with the site key, the action, and the page theme', async () => {
		const fake = installFakeTurnstile();
		document.documentElement.dataset.appearance = 'light';

		const { container } = await render(CaptchaHarness);

		await vi.waitFor(() => {
			expect(fake.options()).toMatchObject({
				sitekey: 'site-key',
				action: 'login',
				theme: 'light',
				size: 'flexible',
				language: 'id'
			});
		});
		await expect.element(page.getByRole('group', { name: 'Verifikasi keamanan' })).toBeVisible();
		await expectAccessible(container);
	});

	it('refreshes the theme and clears the old challenge token', async () => {
		const fake = installFakeTurnstile();
		document.documentElement.dataset.appearance = 'light';
		await render(CaptchaHarness);
		await vi.waitFor(() => fake.options());
		fake.options().callback('old-token');
		await expect.element(page.getByTestId('token')).toHaveTextContent('old-token');
		document.documentElement.dataset.appearance = 'dark';
		window.dispatchEvent(new Event('trustip:appearance'));
		await vi.waitFor(() => {
			expect(fake.options().theme).toBe('dark');
		});
		expect(fake.api.remove).toHaveBeenCalledWith('widget-1');
		await expect.element(page.getByTestId('token')).toHaveTextContent('');
	});

	it('uses the compact widget in narrow forms and adapts after resizing', async () => {
		const fake = installFakeTurnstile();
		const { container } = await render(CaptchaHarness);
		container.style.width = '240px';
		await vi.waitFor(() => {
			expect(fake.options().size).toBe('compact');
		});
		container.style.width = '640px';
		await vi.waitFor(() => {
			expect(fake.options().size).toBe('flexible');
		});
		expect(fake.api.remove).toHaveBeenCalledWith('widget-1');
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
