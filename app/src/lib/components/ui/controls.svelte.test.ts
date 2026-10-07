import { page, userEvent } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Button from './Button.svelte';
import TextField from './TextField.svelte';
import NumberField from './NumberField.svelte';
import MoneyField from './MoneyField.svelte';
import Select from './Select.svelte';
import Switch from './Switch.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';
import { textSnippet } from '#lib/testing/snippet.js';

describe('Button', () => {
	it('renders its label and is accessible', async () => {
		const { container } = await render(Button, { children: textSnippet('Simpan') });

		await expect.element(page.getByRole('button', { name: 'Simpan' })).toBeVisible();
		await expectAccessible(container);
	});

	it('shows Menyimpan and blocks clicks while busy', async () => {
		await render(Button, { children: textSnippet('Simpan'), busy: true });

		const button = page.getByRole('button', { name: 'Menyimpan' });
		await expect.element(button).toBeDisabled();
		await expect.element(button).toHaveAttribute('aria-busy', 'true');
	});

	it('has a touch target of at least 48 pixels', async () => {
		await render(Button, { children: textSnippet('Simpan') });

		const height = page.getByRole('button').element().getBoundingClientRect().height;
		expect(height).toBeGreaterThanOrEqual(48);
	});
});

describe('TextField', () => {
	it('links the label, hint, and error to the input', async () => {
		const { container } = await render(TextField, {
			label: 'Nama usaha',
			hint: 'Nama yang tampil di profil',
			error: 'Nama wajib diisi.'
		});

		const input = page.getByLabelText('Nama usaha');
		await expect.element(input).toHaveAttribute('aria-invalid', 'true');
		await expect
			.element(input)
			.toHaveAccessibleDescription('Nama yang tampil di profil Nama wajib diisi.');
		await expectAccessible(container);
	});
});

describe('NumberField', () => {
	it('steps with large buttons and respects the minimum', async () => {
		const { container } = await render(NumberField, { label: 'Sisa', value: 1 });

		await userEvent.click(page.getByRole('button', { name: 'Kurangi Sisa' }));

		await expect.element(page.getByLabelText('Sisa', { exact: true })).toHaveValue('0');
		await expect.element(page.getByRole('button', { name: 'Kurangi Sisa' })).toBeDisabled();
		await expectAccessible(container);
	});

	it('clamps typed values to the maximum', async () => {
		await render(NumberField, { label: 'Sisa', value: 0, max: 20 });

		await userEvent.fill(page.getByLabelText('Sisa', { exact: true }), '35');
		await userEvent.tab();

		await expect.element(page.getByLabelText('Sisa', { exact: true })).toHaveValue('20');
	});

	it('opens the numeric keypad on phones', async () => {
		await render(NumberField, { label: 'Sisa' });

		await expect
			.element(page.getByLabelText('Sisa', { exact: true }))
			.toHaveAttribute('inputmode', 'numeric');
	});
});

describe('MoneyField', () => {
	it('adds thousand separators while typing', async () => {
		const { container } = await render(MoneyField, { label: 'Harga jual' });

		await userEvent.fill(page.getByLabelText('Harga jual'), '31500');

		await expect.element(page.getByLabelText('Harga jual')).toHaveValue('31.500');
		await expectAccessible(container);
	});

	it('shows an existing amount formatted', async () => {
		await render(MoneyField, { label: 'Harga jual', value: 2000000 });

		await expect.element(page.getByLabelText('Harga jual')).toHaveValue('2.000.000');
	});
});

describe('Select', () => {
	it('is labelled and lists its options', async () => {
		const { container } = await render(Select, {
			label: 'Jadwal setor',
			options: [
				{ value: 'WEEKLY', label: 'Mingguan' },
				{ value: 'MONTHLY', label: 'Bulanan' }
			],
			value: 'WEEKLY'
		});

		await expect.element(page.getByLabelText('Jadwal setor')).toHaveValue('WEEKLY');
		await expectAccessible(container);
	});
});

describe('Switch', () => {
	it('toggles its checked state', async () => {
		const { container } = await render(Switch, { label: 'Tampil di Cari Mitra' });

		await userEvent.click(page.getByRole('switch', { name: 'Tampil di Cari Mitra' }));

		await expect
			.element(page.getByRole('switch', { name: 'Tampil di Cari Mitra' }))
			.toHaveAttribute('aria-checked', 'true');
		await expectAccessible(container);
	});
});
