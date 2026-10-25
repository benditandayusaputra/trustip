<script lang="ts">
	import type { Snippet } from 'svelte';
	import Greeting from '#lib/features/account/components/Greeting.svelte';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import ConsignorDashboard from '#lib/features/dashboard/components/ConsignorDashboard.svelte';
	import DailySalesChart from '#lib/features/reports/components/DailySalesChart.svelte';
	import SplitChart from '#lib/features/reports/components/SplitChart.svelte';
	import type { SalesNarrowing, SalesReport } from '#lib/api/reports.js';
	import { lastThirtyDays } from '#lib/features/dashboard/views.js';
	import { dashboard as id } from '#lib/i18n/dashboard.js';

	const session = getAccountSession();
	const isOwner = $derived(session.account?.membership?.role === 'OWNER');
	const range = lastThirtyDays(new Date());
</script>

<svelte:head>
	<title>{id.dashboard.consignorTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<Greeting />
	{#key isOwner}
		<ConsignorDashboard {isOwner} {range}>
			{#snippet chart(narrowing: SalesNarrowing, controls: Snippet)}
				<DailySalesChart {range} {narrowing} {controls} />
			{/snippet}
			{#snippet split(figures: SalesReport['summary'])}
				<SplitChart {figures} isStore={false} />
			{/snippet}
		</ConsignorDashboard>
	{/key}
</section>
