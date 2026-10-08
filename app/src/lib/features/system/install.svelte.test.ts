import { page, userEvent } from 'vitest/browser';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { noteCompletedAction } from '#lib/offline/install.js';
import InstallHarness from '#lib/testing/InstallHarness.svelte';

const iphone = { isInstalled: () => false, isIos: () => true };
const android = { isInstalled: () => false, isIos: () => false };

function offerInstall() {
	const prompt = vi.fn(() => Promise.resolve());
	const offer = Object.assign(new Event('beforeinstallprompt', { cancelable: true }), {
		prompt,
		userChoice: Promise.resolve({ outcome: 'accepted' })
	});
	window.dispatchEvent(offer);
	return prompt;
}

beforeEach(() => {
	localStorage.clear();
});

describe('Install invite (FR-PWA-03)', () => {
	it('waits until the user has finished one action', async () => {
		await render(InstallHarness, { environment: iphone });

		await expect.element(page.getByRole('complementary')).not.toBeInTheDocument();
		noteCompletedAction();

		await expect.element(page.getByText(/Tambahkan ke Layar Utama/u)).toBeVisible();
	});

	it('offers the browser install prompt after the first action and asks only once', async () => {
		noteCompletedAction();
		await render(InstallHarness, { environment: android });
		const prompt = offerInstall();

		await userEvent.click(page.getByRole('button', { name: 'Pasang' }));

		expect(prompt).toHaveBeenCalledOnce();
		await expect.element(page.getByRole('complementary')).not.toBeInTheDocument();
	});

	it('stays away after the user says later', async () => {
		noteCompletedAction();
		const { unmount } = await render(InstallHarness, { environment: iphone });
		await userEvent.click(page.getByRole('button', { name: 'Nanti' }));
		await unmount();

		await render(InstallHarness, { environment: iphone });

		await expect.element(page.getByRole('complementary')).not.toBeInTheDocument();
	});

	it('says nothing once the app is installed', async () => {
		noteCompletedAction();

		await render(InstallHarness, { environment: { isInstalled: () => true, isIos: () => true } });

		await expect.element(page.getByRole('complementary')).not.toBeInTheDocument();
	});
});
