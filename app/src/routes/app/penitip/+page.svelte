<script lang="ts">
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import ConsignorDashboard from '#lib/features/dashboard/components/ConsignorDashboard.svelte';
	import DailySalesChart from '#lib/features/reports/components/DailySalesChart.svelte';
	import type { SalesNarrowing } from '#lib/api/reports.js';
	import { lastThirtyDays } from '#lib/features/dashboard/views.js';
	import { id } from '#lib/i18n/id.js';

	const session = getAccountSession();
	const isOwner = $derived(session.account?.membership?.role === 'OWNER');
	const range = lastThirtyDays(new Date());
</script>

<svelte:head>
	<title>{id.dashboard.consignorTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">
		{id.account.greeting(session.account?.full_name ?? '')}
	</h1>
	{#key isOwner}
		<ConsignorDashboard {isOwner} {range}>
			{#snippet chart(narrowing: SalesNarrowing)}
				<DailySalesChart {range} {narrowing} />
			{/snippet}
		</ConsignorDashboard>
	{/key}
</section>
