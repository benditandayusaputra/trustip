import { createRawSnippet } from 'svelte';

export function textSnippet(text: string) {
	return createRawSnippet(() => ({ render: () => `<span>${text}</span>` }));
}

export function keySnippet(prefix: string) {
	return createRawSnippet((key: () => string) => ({
		render: () => `<p>${prefix} ${key()}</p>`
	}));
}
