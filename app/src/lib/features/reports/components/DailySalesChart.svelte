<script lang="ts">
	import type { Snippet } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Card from '#lib/components/ui/Card.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { SalesNarrowing } from '#lib/api/reports.js';
	import type { DateRange } from '#lib/domain/ranges.js';
	import {
		dayOfMonth,
		formatBusinessDate,
		formatRupiah,
		formatShortMonth
	} from '#lib/i18n/format.js';
	import { reports as id } from '#lib/i18n/reports.js';
	import { columnChart } from '#lib/domain/chart.js';
	import { dayCells } from '../rows';
	import { SalesReportView } from '../state.svelte';

	interface Props {
		range: DateRange;
		narrowing?: SalesNarrowing;
		controls?: Snippet;
	}

	let { range, narrowing = {}, controls }: Props = $props();

	const text = id.reports;
	const daily = new SalesReportView(apiClient, () => new Date());
	const columns = [
		{ key: 'day' as const, label: text.day },
		{ key: 'sold' as const, label: text.soldUnits, numeric: true },
		{ key: 'gross' as const, label: text.gross, numeric: true }
	];
	let asTable = $state(false);
	const shape = $derived(columnChart(daily.report?.rows.map((row) => row.split.gross) ?? []));
	const points = $derived(
		(daily.report?.rows ?? []).map((row, index) => {
			const tick = dayOfMonth(row.key);
			return {
				key: row.key,
				label: formatBusinessDate(row.key),
				valueLabel: formatRupiah(row.split.gross),
				tick,
				tickNote: index === 0 || tick === '1' ? formatShortMonth(row.key) : ''
			};
		})
	);

	$effect(() => {
		void daily.load(range, 'DAY', narrowing);
	});
</script>

<Card title={text.dailyChart}>
	{#if controls}
		{@render controls()}
	{/if}
	{#if daily.status === 'loading'}
		<Skeleton shape="chart" />
	{:else if daily.status === 'failed' || !daily.report}
		<p class="text-sm text-text">{text.chartFailed}</p>
	{:else}
		{@const report = daily.report}
		{#if asTable}
			<ResponsiveTable caption={text.dailyChart} {columns} rows={dayCells(report)} rowKey="key" />
		{:else}
			{#await import('./ColumnChart.svelte')}
				<Skeleton shape="chart" />
			{:then chart}
				<chart.default
					caption={text.chartCaption(
						formatBusinessDate(report.from),
						formatBusinessDate(report.to)
					)}
					{shape}
					{points}
					ceilingLabel={formatRupiah(shape.ceiling)}
				/>
			{/await}
		{/if}
		<Button variant="secondary" onclick={() => (asTable = !asTable)}>
			{asTable ? text.showChart : text.showTable}
		</Button>
	{/if}
</Card>
