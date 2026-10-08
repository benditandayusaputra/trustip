<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import AccountGate from '#lib/features/account/components/AccountGate.svelte';
	import { onboardingHref, signInHref } from '#lib/features/auth/links.js';
	import NotificationBell from '#lib/features/notifications/components/NotificationBell.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();

	const here = $derived(page.url.pathname + page.url.search);
</script>

<AccountGate
	placement="member"
	onsignedout={() => {
		window.location.assign(signInHref(here));
	}}
	onmisplaced={() => void goto(onboardingHref(here), { replaceState: true })}
>
	{#snippet actions()}
		<NotificationBell />
	{/snippet}
	{@render children()}
</AccountGate>
