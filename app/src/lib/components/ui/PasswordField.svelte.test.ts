import { page, userEvent } from 'vitest/browser';
import { expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PasswordField from './PasswordField.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

it('reveals and hides the same password without losing its value or keyboard focus', async () => {
	const { container } = await render(PasswordField, {
		label: 'Kata sandi',
		autocomplete: 'new-password'
	});
	const input = page.getByLabelText('Kata sandi');
	await userEvent.fill(input, 'Sandi-uji-12345');
	await expect.element(input).toHaveAttribute('type', 'password');
	const toggle = page.getByRole('button', { name: 'Tampilkan sandi' });
	await userEvent.click(toggle);
	await expect.element(input).toHaveAttribute('type', 'text');
	await expect.element(input).toHaveValue('Sandi-uji-12345');
	const hide = page.getByRole('button', { name: 'Sembunyikan sandi' });
	await expect.element(hide).toHaveAttribute('aria-pressed', 'true');
	await expect.element(hide).toHaveFocus();
	await userEvent.click(hide);
	await expect.element(input).toHaveAttribute('type', 'password');
	await expect.element(input).toHaveAttribute('autocomplete', 'new-password');
	await expect.element(input).toHaveValue('Sandi-uji-12345');
	await expectAccessible(container);
});
