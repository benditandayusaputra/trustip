<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import DashboardFigures from './DashboardFigures.svelte';
	import DashboardSkeleton from './DashboardSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { DateRange } from '#lib/domain/ranges.js';
	import { dashboard as id } from '#lib/i18n/dashboard.js';
	import { DashboardState, SalesOverview } from '../state.svelte';
	import { rankedByGross, storeTasks, taskLinksOf } from '../views';
	import InsightList from './InsightList.svelte';
	import OpportunityList from './OpportunityList.svelte';
	import QuickCountList from './QuickCountList.svelte';
	import RankingList from './RankingList.svelte';
	import TaskList from './TaskList.svelte';

	interface Props {
		isOwner: boolean;
		range: DateRange;
	}

	let { isOwner, range }: Props = $props();

	const text = id.dashboard;
	const suggestionPreview = 3;
	const listPreview = 5;
	const board = new DashboardState(apiClient);
	const sales = new SalesOverview(apiClient);
	const summary = $derived(sales.byPartner?.summary);

	$effect(() => {
		void board.load();
		if (isOwner) {
			void sales.load(range);
		}
	});
</script>

{#if board.status === 'loading'}
	<DashboardSkeleton side="STORE" />
{:else if board.status === 'failed' || !board.dashboard}
	<Alert title={text.loadFailed}>{board.failure}</Alert>
{:else}
	{@const figures = board.dashboard.figures}
	<DashboardFigures {figures} {summary} salesStatus={sales.status} {isOwner} side="STORE" />
	<div class="grid items-stretch gap-6 lg:grid-cols-12">
		<div class="min-w-0 lg:col-span-8">
			<QuickCountList partners={board.dashboard.quick_count} previewCount={listPreview} />
		</div>
		<div class="min-w-0 lg:col-span-4">
			<TaskList tasks={taskLinksOf(board.dashboard, storeTasks)} />
		</div>
	</div>
	<InsightList insights={board.insights} side="STORE" previewCount={suggestionPreview} />
	<OpportunityList side="STORE" previewCount={listPreview} />
	{#if isOwner}
		<RankingList
			title={text.rankingProducts}
			rows={rankedByGross(sales.byProduct)}
			status={sales.status}
			previewCount={listPreview}
		/>
	{:else}
		<p class="text-sm text-text-muted">{text.ownerOnly}</p>
	{/if}
{/if}
