<script lang="ts">
	import Card from '#lib/components/ui/Card.svelte';
	import ListDisclosure from '#lib/components/ui/ListDisclosure.svelte';
	import { dashboard as id } from '#lib/i18n/dashboard.js';

	export interface RankedRow {
		key: string;
		label: string;
		value: string;
	}

	interface Props {
		previewCount?: number;
		title: string;
		rows: RankedRow[];
	}

	let { title, rows, previewCount }: Props = $props();
</script>

<Card {title}>
	{#if rows.length === 0}
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
						<li class="flex flex-wrap items-center justify-between gap-4 py-4 text-base text-text">
							<span>{index + 1}. {row.label}</span>
							<span class="font-mono tabular-nums">{row.value}</span>
						</li>
					{/each}
				</ol>
			{/snippet}
		</ListDisclosure>
	{/if}
</Card>
