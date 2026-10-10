<script lang="ts">
	import type { Snippet } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Card from '#lib/components/ui/Card.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { SalesNarrowing } from '#lib/api/reports.js';
	import type { DateRange } from '#lib/domain/ranges.js';
	import { formatBusinessDate, formatRupiah } from '#lib/i18n/format.js';
	import { reports as id } from '#lib/i18n/reports.js';
	import { areaShape } from '#lib/domain/chart.js';
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
	const chartBox = { width: 640, height: 192 };
	let asTable = $state(false);
	const shape = $derived(
		areaShape(daily.report?.rows.map((row) => row.split.gross) ?? [], chartBox)
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
		<Skeleton />
	{:else if daily.status === 'failed' || !daily.report}
		<p class="text-sm text-text">{text.chartFailed}</p>
	{:else}
		{@const report = daily.report}
		{#if asTable}
			<ResponsiveTable caption={text.dailyChart} {columns} rows={dayCells(report)} rowKey="key" />
		{:else}
			{#await import('#lib/components/ui/AreaChart.svelte')}
				<Skeleton />
			{:then chart}
				<chart.default
					caption={text.chartCaption(
						formatBusinessDate(report.from),
						formatBusinessDate(report.to)
					)}
					{shape}
					width={chartBox.width}
					height={chartBox.height}
					ceilingLabel={formatRupiah(shape.ceiling)}
					firstLabel={formatBusinessDate(report.from)}
					lastLabel={formatBusinessDate(report.to)}
				/>
			{/await}
		{/if}
		<Button variant="secondary" onclick={() => (asTable = !asTable)}>
			{asTable ? text.showChart : text.showTable}
		</Button>
	{/if}
</Card>
