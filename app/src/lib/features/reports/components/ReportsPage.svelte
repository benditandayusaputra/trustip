<script lang="ts">
	import type { Snippet } from 'svelte';
	import Tabs from '#lib/components/ui/Tabs.svelte';
	import type { SettlementNarrowing } from '#lib/api/settlements.js';
	import { reports as id } from '#lib/i18n/reports.js';
	import DailyRecapView from './DailyRecapView.svelte';
	import SalesReportView from './SalesReportView.svelte';
	import SettlementReportView from './SettlementReportView.svelte';

	interface Props {
		organizationType: 'STORE' | 'CONSIGNOR';
		settlements: Snippet<[SettlementNarrowing]>;
	}

	let { organizationType, settlements }: Props = $props();

	const text = id.reports;
	const tabs = $derived(
		Object.entries(text.tabs)
			.filter(([key]) => key !== 'daily' || organizationType === 'STORE')
			.map(([key, label]) => ({ key, label }))
	);
	let selected = $state('sales');
	let narrowing = $state<SettlementNarrowing>({ status: '', age: '', query: '' });
</script>

<Tabs label={text.tabsLabel} {tabs} bind:selected param="tab" centered>
	{#snippet panel(key: string)}
		<div class="flex flex-col gap-6">
			{#if key === 'daily'}
				<DailyRecapView />
			{:else if key === 'settlements'}
				<SettlementReportView bind:narrowing />
				{@render settlements(narrowing)}
			{:else}
				<SalesReportView {organizationType} />
			{/if}
		</div>
	{/snippet}
</Tabs>
