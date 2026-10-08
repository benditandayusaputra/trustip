<script lang="ts">
	import type { Snippet } from 'svelte';
	import Alert from '#lib/components/ui/Alert.svelte';
	import KpiCard from '#lib/components/ui/KpiCard.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { SalesNarrowing } from '#lib/api/reports.js';
	import type { DateRange } from '#lib/domain/ranges.js';
	import { formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { DashboardState, SalesOverview } from '../state.svelte';
	import {
		chartChoicesOf,
		consignorTasks,
		narrowingOf,
		rankedByGross,
		taskLinksOf
	} from '../views';
	import InsightList from './InsightList.svelte';
	import OpportunityList from './OpportunityList.svelte';
	import RankingList from './RankingList.svelte';
	import TaskList from './TaskList.svelte';

	interface Props {
		isOwner: boolean;
		range: DateRange;
		chart: Snippet<[SalesNarrowing]>;
	}

	let { isOwner, range, chart }: Props = $props();

	const text = id.dashboard;
	const board = new DashboardState(apiClient);
	const sales = new SalesOverview(apiClient);
	let chartChoice = $state('');
	const summary = $derived(sales.byPartner?.summary);

	$effect(() => {
		void board.load();
		if (isOwner) {
			void sales.load(range);
		}
	});
</script>

{#if board.status === 'loading'}
	<Skeleton />
{:else if board.status === 'failed' || !board.dashboard}
	<Alert title={text.loadFailed}>{board.failure}</Alert>
{:else}
	{@const figures = board.dashboard.figures}
	<div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
		<KpiCard label={text.unitsAtStores} value={formatInteger(figures.stock_units)} />
		{#if summary}
			<KpiCard label={text.soldLast30} value={formatInteger(summary.sold_qty)} />
			<KpiCard
				label={text.consignorShareLast30}
				value={formatRupiah(summary.split.consignor_share)}
			/>
		{/if}
		<KpiCard
			label={text.unpaidSettlements}
			value={formatRupiah(figures.unpaid.amount)}
			caption={text.overdueCaption(figures.overdue.count, formatRupiah(figures.overdue.amount))}
		/>
	</div>
	<div class="grid gap-6 lg:grid-cols-12">
		<div class="flex flex-col gap-6 lg:col-span-4 lg:col-start-9 lg:row-start-1">
			<TaskList tasks={taskLinksOf(board.dashboard, consignorTasks)} />
			<InsightList insights={board.insights} side="CONSIGNOR" />
		</div>
		<div class="flex flex-col gap-4 lg:col-span-8 lg:row-start-1">
			{#if isOwner}
				<Select
					label={text.chartFilter}
					options={chartChoicesOf(sales.byPartner, sales.byProduct)}
					bind:value={chartChoice}
				/>
				{#key chartChoice}
					{@render chart(narrowingOf(chartChoice))}
				{/key}
			{:else}
				<p class="text-sm text-text-muted">{text.ownerOnly}</p>
			{/if}
		</div>
	</div>
	<div class="grid gap-6 md:grid-cols-2">
		{#if isOwner}
			<RankingList title={text.rankingStores} rows={rankedByGross(sales.byPartner)} />
			<RankingList title={text.rankingProducts} rows={rankedByGross(sales.byProduct)} />
		{/if}
		<OpportunityList side="CONSIGNOR" />
	</div>
{/if}
