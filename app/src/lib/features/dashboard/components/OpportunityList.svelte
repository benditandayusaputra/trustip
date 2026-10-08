<script lang="ts">
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import Card from '#lib/components/ui/Card.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { id } from '#lib/i18n/id.js';
	import type { Side } from '../cards';
	import { OpportunityFeed } from '../state.svelte';

	interface Props {
		side: Side;
	}

	let { side }: Props = $props();

	const text = id.dashboard;
	const feed = new OpportunityFeed(apiClient);

	$effect(() => {
		void feed.load(side);
	});
</script>

<Card title={text.opportunitiesTitle}>
	{#if feed.status === 'loading'}
		<Skeleton />
	{:else if feed.status === 'failed'}
		<p class="text-sm text-text-muted">{text.opportunitiesFailed}</p>
	{:else if feed.items.length === 0}
		<p class="text-sm text-text-muted">{text.opportunitiesEmpty}</p>
	{:else}
		<ul class="flex flex-col divide-y divide-border">
			{#each feed.items as item (item.key)}
				<li>
					<a href={item.href} class="flex min-h-12 items-center justify-between gap-4 py-2">
						<span class="flex flex-col">
							<span class="text-base font-semibold text-text">{item.title}</span>
							<span class="text-sm text-text-muted">{item.detail}</span>
						</span>
						<ChevronRight class="size-5 shrink-0 text-text-muted" aria-hidden="true" />
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</Card>
