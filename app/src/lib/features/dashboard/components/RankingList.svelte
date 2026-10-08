<script lang="ts">
	import Card from '#lib/components/ui/Card.svelte';
	import { id } from '#lib/i18n/id.js';

	export interface RankedRow {
		key: string;
		label: string;
		value: string;
	}

	interface Props {
		title: string;
		rows: RankedRow[];
	}

	let { title, rows }: Props = $props();
</script>

<Card {title}>
	{#if rows.length === 0}
		<p class="text-sm text-text-muted">{id.dashboard.rankingEmpty}</p>
	{:else}
		<ol class="flex flex-col divide-y divide-border">
			{#each rows as row, index (row.key)}
				<li class="flex items-center justify-between gap-4 py-2 text-base text-text">
					<span>{index + 1}. {row.label}</span>
					<span class="font-mono tabular-nums">{row.value}</span>
				</li>
			{/each}
		</ol>
	{/if}
</Card>
