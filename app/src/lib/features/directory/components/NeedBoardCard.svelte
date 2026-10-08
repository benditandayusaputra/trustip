<script lang="ts">
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { NeedBoardCard } from '#lib/api/directory.js';
	import { formatBasisPoints, formatDate, formatDistance, formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import ReportListing from './ReportListing.svelte';

	interface Props {
		card: NeedBoardCard;
		canRespond: boolean;
	}

	let { card, canRespond }: Props = $props();

	const text = id.needBoard;
	const respondHref = $derived(
		`/app/ajuan/baru?${new URLSearchParams({ sasaran: card.store.id, kebutuhan: card.id }).toString()}`
	);
</script>

<article class="flex flex-col gap-4 rounded-card border border-border bg-surface p-4">
	<header class="flex flex-col gap-1">
		<h2 class="text-lg font-semibold text-text">{card.title}</h2>
		<p class="text-sm text-text-muted">
			{card.store.name}
			{#if card.store.store_type}
				· {id.storeProfile.storeTypes[card.store.store_type]}
			{/if}
			{#if card.distance_m !== null}
				· <span class="font-mono">{formatDistance(card.distance_m)}</span>
			{/if}
			· {card.store.city_name}
		</p>
	</header>
	<div class="flex flex-wrap gap-2">
		<Badge tone="neutral" label={card.category.name} />
		{#if card.is_own}
			<Badge tone="info" label={text.own} />
		{/if}
	</div>
	{#if card.description}
		<p class="text-base text-text">{card.description}</p>
	{/if}
	<ul class="flex flex-wrap gap-x-4 gap-y-1 text-sm text-text-muted">
		{#if card.est_qty_per_week !== null}
			<li>{text.perWeek(card.est_qty_per_week)}</li>
		{/if}
		{#if card.max_retail_price !== null}
			<li>{text.priceUpTo(formatRupiah(card.max_retail_price))}</li>
		{/if}
		{#if card.offered_commission_bps !== null}
			<li>{text.commission(formatBasisPoints(card.offered_commission_bps))}</li>
		{/if}
		<li>{text.expiresAt(formatDate(card.expires_at))}</li>
	</ul>
	{#if canRespond && !card.is_own}
		<a
			href={respondHref}
			class="inline-flex min-h-11 items-center self-start rounded-control bg-accent px-4 font-semibold text-on-accent"
			>{text.respond}</a
		>
	{/if}
	{#if !card.is_own}
		<ReportListing target={{ target_type: 'NEED', target_id: card.id }} name={card.title} />
	{/if}
</article>
