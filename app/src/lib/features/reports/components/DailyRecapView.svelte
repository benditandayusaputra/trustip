<script lang="ts">
	import CalendarDays from '@lucide/svelte/icons/calendar-days';
	import Alert from '#lib/components/ui/Alert.svelte';
	import DateField from '#lib/components/ui/DateField.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import KpiCard from '#lib/components/ui/KpiCard.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { reports as id } from '#lib/i18n/reports.js';
	import { DailyRecapState } from '../recaps.svelte';
	import { recapCells } from '../rows';

	const text = id.reports;
	const state = new DailyRecapState(apiClient, () => new Date());
	const columns = [
		{ key: 'name' as const, label: text.consignor },
		{ key: 'received' as const, label: text.receivedUnits, numeric: true },
		{ key: 'sold' as const, label: text.soldUnits, numeric: true },
		{ key: 'returned' as const, label: text.returnedUnits, numeric: true },
		{ key: 'gross' as const, label: text.gross, numeric: true },
		{ key: 'storeShare' as const, label: text.storeShare, numeric: true },
		{ key: 'consignorShare' as const, label: text.mustPay, numeric: true },
		{ key: 'paid' as const, label: text.paid, numeric: true }
	];

	$effect(() => {
		void state.load(state.date);
	});
</script>

<DateField label={text.date} value={state.date} onchange={(date: string) => (state.date = date)} />
{#if state.status === 'loading'}
	<ContentSkeleton variant="report" />
{:else if state.status === 'failed' || !state.recap}
	<Alert title={text.loadFailed}>{state.failure}</Alert>
{:else if state.recap.rows.length === 0}
	<EmptyState icon={CalendarDays} message={text.dailyEmpty}>
		{#snippet action()}
			<a href="/app/hitung" class="font-semibold text-accent">{text.toCounts}</a>
		{/snippet}
	</EmptyState>
{:else}
	{@const total = recapCells(state.recap.total)}
	<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
		<KpiCard label={text.gross} value={total.gross} />
		<KpiCard label={text.storeShare} value={total.storeShare} />
		<KpiCard label={text.mustPay} value={total.consignorShare} />
		<KpiCard label={text.paid} value={total.paid} />
	</div>
	<ResponsiveTable
		caption={text.tabs.daily}
		{columns}
		rows={state.recap.rows.map(recapCells)}
		rowKey="key"
	/>
{/if}
