<script lang="ts">
	import Wallet from '@lucide/svelte/icons/wallet';
	import Alert from '#lib/components/ui/Alert.svelte';
	import DateField from '#lib/components/ui/DateField.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import KpiCard from '#lib/components/ui/KpiCard.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { RangePreset } from '#lib/domain/ranges.js';
	import { reports as id } from '#lib/i18n/reports.js';
	import { SettlementReportState } from '../recaps.svelte';
	import { settlementCells } from '../rows';

	const text = id.reports;
	const state = new SettlementReportState(apiClient, () => new Date());
	const presetOptions = Object.entries(text.presets).map(([value, label]) => ({ value, label }));
	const columns = [
		{ key: 'name' as const, label: text.partner },
		{ key: 'issued' as const, label: text.issued, numeric: true },
		{ key: 'paid' as const, label: text.settled, numeric: true },
		{ key: 'unpaid' as const, label: text.unpaid, numeric: true },
		{ key: 'overdue' as const, label: text.overdue, numeric: true }
	];

	$effect(() => {
		void state.load(state.range);
	});
</script>

<fieldset class="grid grid-cols-1 gap-4 md:grid-cols-3">
	<legend class="sr-only">{text.filters}</legend>
	<Select
		label={text.range}
		options={presetOptions}
		value={state.preset}
		onchange={(chosen: string) => (state.preset = chosen as RangePreset)}
	/>
	{#if state.preset === 'CUSTOM'}
		<DateField
			label={text.from}
			value={state.custom.from}
			onchange={(from: string) => (state.custom = { ...state.custom, from })}
		/>
		<DateField
			label={text.to}
			value={state.custom.to}
			onchange={(to: string) => (state.custom = { ...state.custom, to })}
		/>
	{/if}
</fieldset>
{#if state.status === 'loading'}
	<ContentSkeleton variant="report" />
{:else if state.status === 'failed' || !state.report}
	<Alert title={text.loadFailed}>{state.failure}</Alert>
{:else if state.report.rows.length === 0}
	<EmptyState icon={Wallet} message={text.settlementsEmpty}>
		{#snippet action()}
			<a href="/app/setoran" class="font-semibold text-accent">{text.toSettlements}</a>
		{/snippet}
	</EmptyState>
{:else}
	{@const total = settlementCells(state.report.total)}
	<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
		<KpiCard label={text.issued} value={total.issued} />
		<KpiCard label={text.settled} value={total.paid} />
		<KpiCard label={text.unpaid} value={total.unpaid} />
		<KpiCard label={text.overdue} value={total.overdue} />
	</div>
	<ResponsiveTable
		caption={text.tabs.settlements}
		{columns}
		rows={state.report.rows.map(settlementCells)}
		rowKey="key"
	/>
{/if}
