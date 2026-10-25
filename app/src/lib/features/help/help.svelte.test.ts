import { page, userEvent } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import FaqPage from './components/FaqPage.svelte';
import GuidePage from './components/GuidePage.svelte';
import { faq } from '#lib/i18n/faq.js';
import { guide } from '#lib/i18n/guide.js';
import { expectAccessible } from '#lib/testing/accessibility.js';

describe('GuidePage', () => {
	it('walks through every phase in order and links each one from the contents', async () => {
		const { container } = await render(GuidePage, {});

		await expect
			.element(page.getByRole('heading', { level: 1, name: guide.headline }))
			.toBeVisible();
		for (const phase of guide.phases) {
			await expect.element(page.getByRole('heading', { name: phase.title })).toBeVisible();
			await expect
				.element(page.getByRole('navigation', { name: guide.contentsLabel }).getByText(phase.title))
				.toHaveAttribute('href', `#${phase.slug}`);
		}
		await expectAccessible(container);
	});

	it('introduces the three demo personas and leads to the demo sign-in', async () => {
		await render(GuidePage, {});

		for (const name of ['Bendi', 'Zara', 'Jessica']) {
			await expect.element(page.getByText(name, { exact: true })).toBeVisible();
		}
		await expect
			.element(page.getByRole('link', { name: guide.tryDemo }))
			.toHaveAttribute('href', '/demo');
	});
});

describe('FaqPage', () => {
	it('keeps answers folded until the question is opened', async () => {
		const { container } = await render(FaqPage, {});
		const [first] = faq.groups.flatMap((group) => group.items);

		await expect.element(page.getByText(first?.answer ?? '')).not.toBeVisible();
		await userEvent.click(page.getByText(first?.question ?? ''));

		await expect.element(page.getByText(first?.answer ?? '')).toBeVisible();
		await expectAccessible(container);
	});

	it('states honestly that money moves between partners, not through Trustip', () => {
		const answers = faq.groups.flatMap((group) => group.items.map((item) => item.answer));

		expect(answers.join(' ')).toContain('Toko membayar langsung ke penitip');
	});
});
