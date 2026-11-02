<script lang="ts">
	import ChartPie from '@lucide/svelte/icons/chart-pie';
	import Card from '#lib/components/ui/Card.svelte';
	import type { SalesReport } from '#lib/api/reports.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { reports as id } from '#lib/i18n/reports.js';
	import { splitSlices } from '../charts';
	import DonutChart from './DonutChart.svelte';

	interface Props {
		figures: SalesReport['summary'];
		isStore: boolean;
	}

	let { figures, isStore }: Props = $props();

	const text = id.reports;
</script>

<Card title={text.splitTitle} icon={ChartPie} fill>
	<DonutChart
		caption={text.splitCaption}
		slices={splitSlices(figures, isStore)}
		totalLabel={text.gross}
		totalValue={formatRupiah(figures.split.gross)}
	/>
</Card>
