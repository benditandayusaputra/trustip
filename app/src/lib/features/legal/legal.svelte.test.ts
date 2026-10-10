import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import LegalPage from './components/LegalPage.svelte';
import { legal as legalText } from '#lib/i18n/legal.js';
import { expectAccessible } from '#lib/testing/accessibility.js';

const id = { legal: legalText };

describe('LegalPage', () => {
	it('FR-PUB-09 shows the privacy policy with every section heading', async () => {
		const { container } = await render(LegalPage, {
			document: id.legal.privacy,
			related: { href: '/ketentuan', title: id.legal.terms.title }
		});

		await expect
			.element(page.getByRole('heading', { level: 1, name: 'Kebijakan Privasi' }))
			.toBeVisible();
		expect(page.getByRole('heading', { level: 2 }).all()).toHaveLength(
			id.legal.privacy.sections.length
		);
		await expect
			.element(page.getByRole('link', { name: 'Ketentuan Layanan' }))
			.toHaveAttribute('href', '/ketentuan');
		await expectAccessible(container);
	});

	it('R-SEC-13 tells visitors that the captcha widget is used and by whom', () => {
		const mentions = id.legal.privacy.sections.flatMap((section) => section.points).join(' ');

		expect(mentions).toContain('Cloudflare Turnstile');
	});

	it('states honestly what a seal does not prove', () => {
		const terms = id.legal.terms.sections.flatMap((section) => section.points).join(' ');

		expect(terms).toContain('bukan tanda tangan elektronik tersertifikasi');
	});
});
