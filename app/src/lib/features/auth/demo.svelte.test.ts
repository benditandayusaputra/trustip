import { page } from 'vitest/browser';
import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import DemoAccounts from './components/DemoAccounts.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

describe('DemoAccounts', () => {
	it('FR-AUTH-07 shows both demo accounts and hands the chosen one to the login form', async () => {
		const onpick = vi.fn();
		const { container } = await render(DemoAccounts, { onpick });

		await expect.element(page.getByText('toko@demo.trustip.id')).toBeVisible();
		await page.getByRole('button', { name: 'Isi sebagai toko (Warung Bu Sri)' }).click();

		expect(onpick).toHaveBeenCalledWith(expect.objectContaining({ email: 'toko@demo.trustip.id' }));
		await expectAccessible(container);
	});
});
