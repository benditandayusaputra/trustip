<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import KpiCard from '#lib/components/ui/KpiCard.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { DateRange } from '#lib/domain/ranges.js';
	import { formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { DashboardState, SalesOverview } from '../state.svelte';
	import { rankedByGross, storeTasks, taskLinksOf } from '../views';
	import InsightList from './InsightList.svelte';
	import QuickCountList from './QuickCountList.svelte';
	import RankingList from './RankingList.svelte';
	import TaskList from './TaskList.svelte';

	interface Props {
		isOwner: boolean;
		range: DateRange;
	}

	let { isOwner, range }: Props = $props();

	const text = id.dashboard;
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
	<Skeleton />
{:else if board.status === 'failed' || !board.dashboard}
	<Alert title={text.loadFailed}>{board.failure}</Alert>
{:else}
	{@const figures = board.dashboard.figures}
	<div class="grid grid-cols-2 gap-4 lg:grid-cols-5">
		<KpiCard label={text.activePartners} value={formatInteger(figures.active_partners)} />
		<KpiCard label={text.unitsConsigned} value={formatInteger(figures.stock_units)} />
		{#if summary}
			<KpiCard label={text.grossLast30} value={formatRupiah(summary.split.gross)} />
			<KpiCard label={text.storeShareLast30} value={formatRupiah(summary.split.store_share)} />
		{/if}
		<KpiCard
			label={text.mustPayTotal}
			value={formatRupiah(figures.unpaid.amount)}
			caption={text.dueSoonCaption(figures.due_soon.count, formatRupiah(figures.due_soon.amount))}
		/>
	</div>
	<div class="grid gap-6 lg:grid-cols-12">
		<div class="flex flex-col gap-6 lg:col-span-8">
			<QuickCountList partners={board.dashboard.quick_count} />
			{#if isOwner}
				<RankingList title={text.rankingProducts} rows={rankedByGross(sales.byProduct)} />
			{:else}
				<p class="text-sm text-text-muted">{text.ownerOnly}</p>
			{/if}
		</div>
		<div class="flex flex-col gap-6 lg:col-span-4">
			<TaskList tasks={taskLinksOf(board.dashboard, storeTasks)} />
			<InsightList insights={board.insights} />
		</div>
	</div>
{/if}
