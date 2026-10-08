<script lang="ts">
	import Card from '#lib/components/ui/Card.svelte';
	import type { Dashboard } from '#lib/api/insights.js';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		partners: Dashboard['quick_count'];
	}

	let { partners }: Props = $props();

	const text = id.dashboard;
</script>

<Card title={text.quickCountTitle}>
	{#if partners.length === 0}
		<p class="text-sm text-text-muted">{text.quickCountEmpty}</p>
	{:else}
		<ul class="flex flex-col divide-y divide-border">
			{#each partners as partner (partner.partnership_id)}
				<li class="flex items-center justify-between gap-4 py-2">
					<span class="flex flex-col">
						<span class="text-base font-semibold text-text">{partner.counterpart_name}</span>
						<span class="text-sm text-text-muted">{text.quickCountRow(partner.stock_units)}</span>
					</span>
					<a
						href="/app/hitung/baru?mitra={partner.partnership_id}"
						class="inline-flex min-h-12 items-center rounded-control bg-accent px-4 font-semibold text-on-accent"
						aria-label="{text.count} {partner.counterpart_name}"
					>
						{text.count}
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</Card>
