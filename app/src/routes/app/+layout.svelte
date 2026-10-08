<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import AccountGate from '#lib/features/account/components/AccountGate.svelte';
	import { onboardingHref, signInHref } from '#lib/features/auth/links.js';
	import NotificationBell from '#lib/features/notifications/components/NotificationBell.svelte';
	import OfflineKeeper from '#lib/features/outbox/components/OfflineKeeper.svelte';
	import OutboxBadge from '#lib/features/outbox/components/OutboxBadge.svelte';
	import InstallPrompt from '#lib/features/system/InstallPrompt.svelte';
	import OfflineRibbon from '#lib/features/system/OfflineRibbon.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();

	const here = $derived(page.url.pathname + page.url.search);
</script>

<OfflineRibbon />
<InstallPrompt />
<AccountGate
	placement="member"
	onsignedout={() => {
		window.location.assign(signInHref(here));
	}}
	onmisplaced={() => void goto(onboardingHref(here), { replaceState: true })}
>
	{#snippet actions()}
		<OfflineKeeper />
		<OutboxBadge />
		<NotificationBell />
	{/snippet}
	{@render children()}
</AccountGate>
