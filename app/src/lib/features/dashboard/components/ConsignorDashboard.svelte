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
	import { dashboard as id } from '#lib/i18n/dashboard.js';
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
		chart: Snippet<[SalesNarrowing, Snippet]>;
	}

	let { isOwner, range, chart }: Props = $props();

	const text = id.dashboard;
	const suggestionPreview = 3;
	const listPreview = 5;
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
	<div class="grid grid-cols-2 gap-4 {summary ? 'lg:grid-cols-4' : ''}">
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
	{#snippet chartControls()}
		<Select
			label={text.chartFilter}
			options={chartChoicesOf(sales.byPartner, sales.byProduct)}
			bind:value={chartChoice}
		/>
	{/snippet}
	{#if isOwner}
		<div class="grid items-start gap-6 lg:grid-cols-12">
			<div class="lg:col-span-4 lg:col-start-9 lg:row-start-1">
				<TaskList tasks={taskLinksOf(board.dashboard, consignorTasks)} />
			</div>
			<div class="min-w-0 lg:col-span-8 lg:row-start-1">
				{@render chart(narrowingOf(chartChoice), chartControls)}
			</div>
		</div>
	{:else}
		<TaskList tasks={taskLinksOf(board.dashboard, consignorTasks)} />
		<p class="text-sm text-text-muted">{text.ownerOnly}</p>
	{/if}
	<InsightList insights={board.insights} side="CONSIGNOR" previewCount={suggestionPreview} />
	{#if isOwner}
		<div class="grid items-start gap-6 lg:grid-cols-2">
			<RankingList
				title={text.rankingStores}
				rows={rankedByGross(sales.byPartner)}
				previewCount={listPreview}
			/>
			<RankingList
				title={text.rankingProducts}
				rows={rankedByGross(sales.byProduct)}
				previewCount={listPreview}
			/>
		</div>
	{/if}
	<OpportunityList side="CONSIGNOR" previewCount={listPreview} />
{/if}
