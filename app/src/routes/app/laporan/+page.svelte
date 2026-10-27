<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import ChartColumn from '@lucide/svelte/icons/chart-column';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import ReportsPage from '#lib/features/reports/components/ReportsPage.svelte';
	import SettlementList from '#lib/features/settlements/components/SettlementList.svelte';
	import { reports as id } from '#lib/i18n/reports.js';

	const session = getAccountSession();
	const organizationType = $derived(
		session.account?.membership?.organization_type === 'CONSIGNOR' ? 'CONSIGNOR' : 'STORE'
	);
</script>

<svelte:head>
	<title>{id.reports.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<PageHeading icon={ChartColumn}>{id.reports.title}</PageHeading>
	{#key organizationType}
		<ReportsPage {organizationType}>
			{#snippet settlements()}<SettlementList />{/snippet}
		</ReportsPage>
	{/key}
</section>
