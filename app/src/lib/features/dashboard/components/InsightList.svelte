<script lang="ts">
	import Badge from '#lib/components/ui/Badge.svelte';
	import Card from '#lib/components/ui/Card.svelte';
	import { edgeTone } from '#lib/components/ui/tone.js';
	import type { Insight } from '#lib/api/insights.js';
	import { dashboard as id } from '#lib/i18n/dashboard.js';
	import { cardOf, type Side } from '../cards';

	interface Props {
		insights: Insight[];
		side: Side;
	}

	let { insights, side }: Props = $props();

	const cards = $derived(insights.map((insight, index) => cardOf(insight, index, side)));
</script>

<Card title={id.dashboard.insightsTitle}>
	{#if cards.length === 0}
		<p class="text-sm text-text-muted">{id.dashboard.noInsights}</p>
	{:else}
		<ul class="flex flex-col gap-4">
			{#each cards as card (card.key)}
				<li
					class="flex flex-col gap-2 rounded-control border border-l-4 border-border p-4 {edgeTone[
						card.tone
					]}"
				>
					<Badge tone={card.tone} label={card.level} />
					<p class="text-base font-semibold text-text">{card.title}</p>
					<p class="text-sm text-text-muted">{card.body}</p>
					<a href={card.href} class="inline-flex min-h-11 items-center font-semibold text-accent">
						{card.actionLabel}
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</Card>
