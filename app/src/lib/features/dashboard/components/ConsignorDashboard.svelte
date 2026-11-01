<script lang="ts">
	import type { Snippet } from 'svelte';
	import Alert from '#lib/components/ui/Alert.svelte';
	import DashboardFigures from './DashboardFigures.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import DashboardSkeleton from '#lib/components/ui/DashboardSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { SalesNarrowing, SalesReport } from '#lib/api/reports.js';
	import type { DateRange } from '#lib/domain/ranges.js';
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
	import RankingList from './RankingList.svelte';
	import TaskList from './TaskList.svelte';

	interface Props {
		isOwner: boolean;
		range: DateRange;
		chart: Snippet<[SalesNarrowing, Snippet]>;
		split: Snippet<[SalesReport['summary']]>;
	}

	let { isOwner, range, chart, split }: Props = $props();

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
	<DashboardSkeleton side="CONSIGNOR" />
{:else if board.status === 'failed' || !board.dashboard}
	<Alert title={text.loadFailed}>{board.failure}</Alert>
{:else}
	{@const figures = board.dashboard.figures}
	<DashboardFigures {figures} {summary} salesStatus={sales.status} {isOwner} side="CONSIGNOR" />
	{#snippet chartControls()}
		<Select
			label={text.chartFilter}
			options={chartChoicesOf(sales.byPartner, sales.byProduct)}
			bind:value={chartChoice}
		/>
	{/snippet}
	{#if isOwner}
		<div class="grid items-stretch gap-6 lg:grid-cols-12">
			<div class="flex flex-col gap-6 lg:col-span-4 lg:col-start-9 lg:row-start-1">
				<TaskList tasks={taskLinksOf(board.dashboard, consignorTasks)} />
				{#if summary}{@render split(summary)}{/if}
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
		<div class="flex flex-col gap-6">
			<RankingList
				title={text.rankingStores}
				rows={rankedByGross(sales.byPartner)}
				status={sales.status}
				previewCount={listPreview}
			/>
			<RankingList
				title={text.rankingProducts}
				rows={rankedByGross(sales.byProduct)}
				status={sales.status}
				previewCount={listPreview}
			/>
		</div>
	{/if}
{/if}
