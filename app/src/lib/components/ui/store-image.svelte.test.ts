import { expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import StoreImage from './StoreImage.svelte';

it('shows the default store picture when the store has no photo', async () => {
	const { container } = await render(StoreImage, { url: null });

	expect(container.querySelector('[data-default-store-image]')).not.toBeNull();
	expect(container.querySelector('img')).toBeNull();
});

it('swaps to the default store picture when the photo cannot be loaded', async () => {
	const { container } = await render(StoreImage, { url: '/foto-toko-tidak-ada.jpg' });

	await expect
		.poll(() => container.querySelector('[data-default-store-image]'), { timeout: 5000 })
		.not.toBeNull();
	expect(container.querySelector('img')).toBeNull();
});
