<script lang="ts">
	import ChartBar from '@lucide/svelte/icons/chart-bar';
	import ChartBarStacked from '@lucide/svelte/icons/chart-bar-stacked';
	import Card from '#lib/components/ui/Card.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { DateRange } from '#lib/domain/ranges.js';
	import { reports as id } from '#lib/i18n/reports.js';
	import { grossBars, returnBars, returnLegend } from '../charts';
	import { SalesReportView } from '../state.svelte';
	import BarChart from './BarChart.svelte';
	import SplitChart from './SplitChart.svelte';

	interface Props {
		range: DateRange;
		isStore: boolean;
	}

	let { range, isStore }: Props = $props();

	const text = id.reports;
	const partners = new SalesReportView(apiClient, () => new Date());
	const products = new SalesReportView(apiClient, () => new Date());
	const status = $derived(
		[partners.status, products.status].find((each) => each !== 'ready') ?? 'ready'
	);

	$effect(() => {
		void partners.load(range, 'PARTNER');
		void products.load(range, 'PRODUCT');
	});
</script>

{#if status === 'loading' && (!partners.report || !products.report)}
	<div class="grid gap-6 lg:grid-cols-2">
		<Skeleton shape="chart" />
		<Skeleton shape="chart" />
	</div>
{:else if status === 'failed' || !partners.report || !products.report}
	<p role="status" class="text-sm text-text-muted">{text.chartsFailed}</p>
{:else}
	{@const byPartner = partners.report}
	{@const byProduct = products.report}
	<div class="grid items-stretch gap-6 lg:grid-cols-12">
		<div class="min-w-0 lg:col-span-5">
			<SplitChart figures={byPartner.summary} {isStore} />
		</div>
		<div class="min-w-0 lg:col-span-7">
			<Card title={text.partnersChart[isStore ? 'STORE' : 'CONSIGNOR']} icon={ChartBar} fill>
				<BarChart caption={text.partnersCaption} rows={grossBars(byPartner)} />
			</Card>
		</div>
	</div>
	<Card title={text.productsChart} icon={ChartBarStacked}>
		<BarChart caption={text.productsCaption} rows={returnBars(byProduct)} legend={returnLegend} />
	</Card>
{/if}
