<script lang="ts">
	import ChartColumn from '@lucide/svelte/icons/chart-column';
	import Alert from '#lib/components/ui/Alert.svelte';
	import DateField from '#lib/components/ui/DateField.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import KpiCard from '#lib/components/ui/KpiCard.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { ReportGrouping } from '#lib/api/reports.js';
	import type { RangePreset } from '#lib/domain/ranges.js';
	import { reports as id } from '#lib/i18n/reports.js';
	import Button from '#lib/components/ui/Button.svelte';
	import { downloadCsv } from '../download';
	import { reportRows, salesCsv, summaryCards } from '../rows';
	import { SalesReportView } from '../state.svelte';
	import DailySalesChart from './DailySalesChart.svelte';

	interface Props {
		organizationType: 'STORE' | 'CONSIGNOR';
	}

	let { organizationType }: Props = $props();

	const text = id.reports;
	const view = new SalesReportView(apiClient, () => new Date());
	const isStore = $derived(organizationType === 'STORE');
	const presetOptions = Object.entries(text.presets).map(([value, label]) => ({ value, label }));
	const groupOptions = $derived(
		Object.entries(text.groupings[organizationType]).map(([value, label]) => ({ value, label }))
	);
	const columns = $derived([
		{ key: 'label' as const, label: text.rowLabel },
		{ key: 'sold' as const, label: text.soldUnits, numeric: true },
		{ key: 'gross' as const, label: text.gross, numeric: true },
		{ key: 'storeShare' as const, label: text.storeShare, numeric: true },
		{
			key: 'consignorShare' as const,
			label: isStore ? text.mustPay : text.consignorShare,
			numeric: true
		},
		{ key: 'returned' as const, label: text.returnedUnits, numeric: true },
		{ key: 'returnedPercent' as const, label: text.returnedPercent, numeric: true }
	]);

	$effect(() => {
		void view.load(view.range, view.groupBy);
	});
</script>

<fieldset class="grid grid-cols-1 gap-4 md:grid-cols-2">
	<legend class="sr-only">{text.filters}</legend>
	<Select
		label={text.range}
		options={presetOptions}
		value={view.preset}
		onchange={(chosen: string) => (view.preset = chosen as RangePreset)}
	/>
	<Select
		label={text.groupBy}
		options={groupOptions}
		value={view.groupBy}
		onchange={(chosen: string) => (view.groupBy = chosen as ReportGrouping)}
	/>
	{#if view.preset === 'CUSTOM'}
		<DateField
			label={text.from}
			value={view.custom.from}
			onchange={(from: string) => (view.custom = { ...view.custom, from })}
		/>
		<DateField
			label={text.to}
			value={view.custom.to}
			onchange={(to: string) => (view.custom = { ...view.custom, to })}
		/>
	{/if}
</fieldset>
{#if view.status === 'loading'}
	<ContentSkeleton variant="report" />
{:else if view.status === 'failed' || !view.report}
	<Alert title={text.loadFailed}>{view.failure}</Alert>
{:else}
	{@const report = view.report}
	<div class="grid grid-cols-2 gap-4 md:grid-cols-3">
		{#each summaryCards(report.summary, isStore) as card (card.label)}
			<KpiCard label={card.label} value={card.value} />
		{/each}
	</div>
	{#if report.summary.sold_qty === 0 && report.summary.returned_qty === 0}
		<EmptyState icon={ChartColumn} message={text.empty}>
			{#snippet action()}
				<a href="/app/hitung" class="font-semibold text-accent">{text.toCounts}</a>
			{/snippet}
		</EmptyState>
	{:else}
		<DailySalesChart range={view.range} />
		<ResponsiveTable caption={text.salesTitle} {columns} rows={reportRows(report)} rowKey="key" />
		<Button
			variant="secondary"
			onclick={() => {
				downloadCsv(text.fileName('penjualan', report.from, report.to), salesCsv(report, isStore));
			}}
		>
			{text.exportCsv}
		</Button>
	{/if}
{/if}
