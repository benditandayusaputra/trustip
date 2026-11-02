<script lang="ts">
	import Greeting from '#lib/features/account/components/Greeting.svelte';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import StoreDashboard from '#lib/features/dashboard/components/StoreDashboard.svelte';
	import DailySalesChart from '#lib/features/reports/components/DailySalesChart.svelte';
	import SplitChart from '#lib/features/reports/components/SplitChart.svelte';
	import type { SalesReport } from '#lib/api/reports.js';
	import { lastThirtyDays } from '#lib/features/dashboard/views.js';
	import { dashboard as id } from '#lib/i18n/dashboard.js';

	const session = getAccountSession();
	const isOwner = $derived(session.account?.membership?.role === 'OWNER');
	const range = lastThirtyDays(new Date());
</script>

<svelte:head>
	<title>{id.dashboard.storeTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<Greeting />
	{#key isOwner}
		<StoreDashboard {isOwner} {range}>
			{#snippet chart(daily: SalesReport | undefined)}
				<DailySalesChart {range} initial={daily} />
			{/snippet}
			{#snippet split(figures: SalesReport['summary'])}
				<SplitChart {figures} isStore />
			{/snippet}
		</StoreDashboard>
	{/key}
</section>
