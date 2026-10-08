<script lang="ts">
	import Tabs from '#lib/components/ui/Tabs.svelte';
	import { id } from '#lib/i18n/id.js';
	import DailyRecapView from './DailyRecapView.svelte';
	import SalesReportView from './SalesReportView.svelte';
	import SettlementReportView from './SettlementReportView.svelte';

	interface Props {
		organizationType: 'STORE' | 'CONSIGNOR';
	}

	let { organizationType }: Props = $props();

	const text = id.reports;
	const tabs = $derived(
		Object.entries(text.tabs)
			.filter(([key]) => key !== 'daily' || organizationType === 'STORE')
			.map(([key, label]) => ({ key, label }))
	);
	let selected = $state('sales');
</script>

<Tabs label={text.tabsLabel} {tabs} bind:selected>
	{#snippet panel(key: string)}
		<div class="flex flex-col gap-6">
			{#if key === 'daily'}
				<DailyRecapView />
			{:else if key === 'settlements'}
				<SettlementReportView />
			{:else}
				<SalesReportView {organizationType} />
			{/if}
		</div>
	{/snippet}
</Tabs>
