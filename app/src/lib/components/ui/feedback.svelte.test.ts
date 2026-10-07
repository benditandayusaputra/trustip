import { page, userEvent } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Inbox from '@lucide/svelte/icons/inbox';
import Card from './Card.svelte';
import KpiCard from './KpiCard.svelte';
import Badge from './Badge.svelte';
import Alert from './Alert.svelte';
import Dialog from './Dialog.svelte';
import BottomSheet from './BottomSheet.svelte';
import Skeleton from './Skeleton.svelte';
import EmptyState from './EmptyState.svelte';
import Tabs from './Tabs.svelte';
import ResponsiveTable from './ResponsiveTable.svelte';
import { Toaster } from './toast.svelte';
import ToasterHarness from '#lib/testing/ToasterHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';
import { keySnippet, textSnippet } from '#lib/testing/snippet.js';

describe('Card and KpiCard', () => {
	it('render a titled card and a key figure', async () => {
		const card = await render(Card, { title: 'Penjualan', children: textSnippet('Isi') });
		const kpi = await render(KpiCard, { label: 'Harus setor', value: 'Rp31.500' });

		await expect.element(page.getByRole('heading', { name: 'Penjualan' })).toBeVisible();
		await expect.element(page.getByText('Rp31.500')).toBeVisible();
		await expectAccessible(card.container);
		await expectAccessible(kpi.container);
	});
});

describe('Badge', () => {
	it('always carries a word, not only a color', async () => {
		const { container } = await render(Badge, { tone: 'success', label: 'Lunas' });

		await expect.element(page.getByText('Lunas')).toBeVisible();
		await expectAccessible(container);
	});
});

describe('Alert', () => {
	it('shows a title, body, and one action', async () => {
		const { container } = await render(Alert, {
			title: 'Setoran lewat tempo',
			children: textSnippet('Setoran ST-202610-0003 belum dibayar.'),
			action: textSnippet('Lihat Setoran')
		});

		await expect.element(page.getByText('Setoran lewat tempo')).toBeVisible();
		await expectAccessible(container);
	});
});

describe('Toaster', () => {
	it('announces toasts politely and dismisses them', async () => {
		const toaster = new Toaster();
		const { container } = await render(ToasterHarness, { toaster });

		toaster.show({ tone: 'success', message: 'Titipan masuk tersimpan.' });

		const region = page.getByRole('region', { name: 'Pemberitahuan' });
		await expect.element(region).toHaveAttribute('aria-live', 'polite');
		await expect.element(page.getByText('Titipan masuk tersimpan.')).toBeVisible();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Tutup pemberitahuan' }));
		await expect.element(page.getByText('Titipan masuk tersimpan.')).not.toBeInTheDocument();
	});

	it('keeps at most three toasts', async () => {
		const toaster = new Toaster();
		await render(ToasterHarness, { toaster });

		['satu', 'dua', 'tiga', 'empat'].forEach((message) => {
			toaster.show({ tone: 'info', message });
		});

		expect(toaster.items.map((toast) => toast.message)).toEqual(['dua', 'tiga', 'empat']);
	});
});

describe('Dialog and BottomSheet', () => {
	it('opens as a modal and closes with Escape', async () => {
		await render(Dialog, { open: true, title: 'Batalkan titipan?', children: textSnippet('Isi') });

		await expect.element(page.getByRole('dialog', { name: 'Batalkan titipan?' })).toBeVisible();
		await userEvent.keyboard('{Escape}');
		await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();
	});

	it('renders the bottom sheet as an accessible dialog', async () => {
		const { container } = await render(BottomSheet, {
			open: true,
			title: 'Pilih mitra',
			children: textSnippet('Daftar')
		});

		await expect.element(page.getByRole('dialog', { name: 'Pilih mitra' })).toBeVisible();
		await expectAccessible(container);
	});
});

describe('Skeleton and EmptyState', () => {
	it('labels the loading skeleton for screen readers', async () => {
		const { container } = await render(Skeleton, { lines: 3 });

		await expect.element(page.getByRole('status', { name: 'Memuat' })).toBeInTheDocument();
		await expectAccessible(container);
	});

	it('shows an icon, one sentence, and one action', async () => {
		const { container } = await render(EmptyState, {
			icon: Inbox,
			message: 'Belum ada titipan masuk.',
			action: textSnippet('Catat titipan')
		});

		await expect.element(page.getByText('Belum ada titipan masuk.')).toBeVisible();
		await expectAccessible(container);
	});
});

describe('Tabs', () => {
	it('moves between tabs with the arrow keys', async () => {
		const { container } = await render(Tabs, {
			label: 'Halaman mitra',
			tabs: [
				{ key: 'ringkasan', label: 'Ringkasan' },
				{ key: 'stok', label: 'Stok' }
			],
			panel: keySnippet('Panel')
		});

		await userEvent.click(page.getByRole('tab', { name: 'Ringkasan' }));
		await userEvent.keyboard('{ArrowRight}');

		await expect
			.element(page.getByRole('tab', { name: 'Stok' }))
			.toHaveAttribute('aria-selected', 'true');
		await expect.element(page.getByRole('tabpanel', { name: 'Stok' })).toBeVisible();
		await expectAccessible(container);
	});
});

describe('ResponsiveTable', () => {
	const props = {
		caption: 'Stok per produk',
		columns: [
			{ key: 'product', label: 'Produk' },
			{ key: 'qty', label: 'Stok', numeric: true }
		],
		rows: [
			{ product: 'Roti cokelat', qty: '18' },
			{ product: 'Keripik pedas', qty: '6' }
		],
		rowKey: 'product'
	};

	it('shows a table on wide screens', async () => {
		await page.viewport(1024, 800);
		const { container } = await render(ResponsiveTable, props);

		await expect.element(page.getByRole('table')).toBeVisible();
		await expectAccessible(container);
	});

	it('turns rows into cards below 768 pixels', async () => {
		await page.viewport(375, 800);
		const { container } = await render(ResponsiveTable, props);

		await expect.element(page.getByRole('table')).not.toBeInTheDocument();
		await expect.element(page.getByRole('list', { name: 'Stok per produk' })).toBeVisible();
		await expectAccessible(container);
	});
});
