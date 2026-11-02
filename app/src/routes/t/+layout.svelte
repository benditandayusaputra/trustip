<script lang="ts">
	import { PUBLIC_STREAM_URL } from '$app/env/public';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import AccountGate from '#lib/features/account/components/AccountGate.svelte';
	import { onboardingHref, signInHref } from '#lib/features/auth/links.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();

	const here = $derived(page.url.pathname + page.url.search);
</script>

<AccountGate
	placement="anyone"
	onsignedout={() => {
		window.location.assign(signInHref(here));
	}}
	onmisplaced={() => void goto(onboardingHref(here), { replaceState: true })}
>
	{#snippet actions()}
		{#await import('#lib/features/notifications/components/NotificationBell.svelte') then { default: NotificationBell }}
			<NotificationBell streamUrl={PUBLIC_STREAM_URL} />
		{/await}
	{/snippet}
	{@render children()}
</AccountGate>
