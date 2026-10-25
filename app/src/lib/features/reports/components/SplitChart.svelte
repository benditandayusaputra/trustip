<script lang="ts">
	import ChartPie from '@lucide/svelte/icons/chart-pie';
	import Card from '#lib/components/ui/Card.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import type { SalesReport } from '#lib/api/reports.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { reports as id } from '#lib/i18n/reports.js';
	import { splitSlices } from '../charts';

	interface Props {
		figures: SalesReport['summary'];
		isStore: boolean;
	}

	let { figures, isStore }: Props = $props();

	const text = id.reports;
</script>

<Card title={text.splitTitle} icon={ChartPie} fill>
	{#await import('./DonutChart.svelte')}
		<Skeleton shape="chart" />
	{:then chart}
		<chart.default
			caption={text.splitCaption}
			slices={splitSlices(figures, isStore)}
			totalLabel={text.gross}
			totalValue={formatRupiah(figures.split.gross)}
		/>
	{/await}
</Card>
