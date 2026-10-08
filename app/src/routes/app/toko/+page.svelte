<script lang="ts">
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import StoreDashboard from '#lib/features/dashboard/components/StoreDashboard.svelte';
	import { lastThirtyDays } from '#lib/features/dashboard/views.js';
	import { id } from '#lib/i18n/id.js';

	const session = getAccountSession();
	const isOwner = $derived(session.account?.membership?.role === 'OWNER');
	const range = lastThirtyDays(new Date());
</script>

<svelte:head>
	<title>{id.dashboard.storeTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">
		{id.account.greeting(session.account?.full_name ?? '')}
	</h1>
	{#key isOwner}
		<StoreDashboard {isOwner} {range} />
	{/key}
</section>
