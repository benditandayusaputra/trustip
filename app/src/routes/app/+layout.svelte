<script lang="ts">
	import { PUBLIC_STREAM_URL } from '$app/env/public';
	import { goto } from '$app/navigation';
	import type { Account } from '#lib/api/users.js';
	import { page } from '$app/state';
	import { dashboardPath } from '#lib/features/account/session.svelte.js';
	import DashboardSkeleton from '#lib/features/dashboard/components/DashboardSkeleton.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import AccountGate from '#lib/features/account/components/AccountGate.svelte';
	import { onboardingHref, signInHref } from '#lib/features/auth/links.js';
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
	onmisplaced={(account: Account) =>
		void goto(
			account.membership
				? dashboardPath(account.membership.organization_type)
				: onboardingHref(here),
			{ replaceState: true }
		)}
>
	{#snippet loading()}
		{#if page.url.pathname === '/app/toko' || page.url.pathname === '/app/penitip'}
			<DashboardSkeleton side={page.url.pathname === '/app/toko' ? 'STORE' : 'CONSIGNOR'} />
		{:else}
			<Skeleton />
		{/if}
	{/snippet}
	{#snippet actions()}
		<OfflineKeeper />
		<OutboxBadge />
		{#await import('#lib/features/notifications/components/NotificationBell.svelte') then { default: NotificationBell }}
			<NotificationBell streamUrl={PUBLIC_STREAM_URL} />
		{/await}
	{/snippet}
	{@render children()}
</AccountGate>
