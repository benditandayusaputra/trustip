<script lang="ts">
	import ChartNoAxesCombined from '@lucide/svelte/icons/chart-no-axes-combined';
	import RankingEntry from './RankingEntry.svelte';
	import Card from '#lib/components/ui/Card.svelte';
	import ListDisclosure from '#lib/components/ui/ListDisclosure.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { safeHref } from '#lib/domain/safe-href.js';
	import { dashboard as id } from '#lib/i18n/dashboard.js';

	import type { RankedRow } from '../ranking';

	interface Props {
		previewCount?: number;
		title: string;
		rows: RankedRow[];
		status?: 'loading' | 'ready' | 'failed';
	}
	let { title, rows, previewCount, status = 'ready' }: Props = $props();
	const loadingRows = 5;
</script>

<Card icon={ChartNoAxesCombined} {title}>
	{#if status === 'loading'}
		<Skeleton lines={loadingRows} shape="control" />
	{:else if status === 'failed'}
		<p class="text-sm text-text-muted">{id.dashboard.salesFailed}</p>
	{:else if rows.length === 0}
		<p class="text-sm text-text-muted">{id.dashboard.rankingEmpty}</p>
	{:else}
		<ListDisclosure
			count={rows.length}
			{previewCount}
			moreLabel={id.dashboard.allRows(rows.length)}
			lessLabel={id.dashboard.fewerRows}
		>
			{#snippet children(limit: number)}
				<ol class="flex flex-col divide-y divide-border">
					{#each rows.slice(0, limit) as row, index (row.key)}
						<li>
							{#if row.product}
								<a
									href={safeHref(row.product.href)}
									class="flex flex-wrap items-center justify-between gap-4 rounded-control py-4 text-base text-text hover:bg-surface-hover"
								>
									<RankingEntry {row} {index} />
								</a>
							{:else}
								<div
									class="flex flex-wrap items-center justify-between gap-4 py-4 text-base text-text"
								>
									<RankingEntry {row} {index} />
								</div>
							{/if}
						</li>
					{/each}
				</ol>
			{/snippet}
		</ListDisclosure>
	{/if}
</Card>
