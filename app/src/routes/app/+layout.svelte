<script lang="ts">
	import { page } from '$app/state';
	import AccountGate from '#lib/features/account/components/AccountGate.svelte';
	import { signInHref } from '#lib/features/auth/links.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<AccountGate
	onsignedout={() => {
		window.location.assign(signInHref(page.url.pathname + page.url.search));
	}}
>
	{@render children()}
</AccountGate>
