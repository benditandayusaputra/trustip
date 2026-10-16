<script lang="ts">
	import Megaphone from '@lucide/svelte/icons/megaphone';
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { NeedBoardCard } from '#lib/api/directory.js';
	import { formatBasisPoints, formatDate, formatDistance, formatRupiah } from '#lib/i18n/format.js';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { organization as organizationText } from '#lib/i18n/organization.js';
	import LazyReportListing from './LazyReportListing.svelte';

	const id = { ...directoryText, ...organizationText };

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

<article class="surface-card flex flex-col gap-4 p-5">
	<Megaphone class="size-6 text-accent" aria-hidden="true" />
	<header class="flex flex-col gap-1">
		<h2 class="text-lg font-semibold text-text">
			<a href="/kebutuhan/{card.id}" class="hover:text-accent">{card.title}</a>
		</h2>
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
			class="inline-flex min-h-11 items-center self-start rounded-control bg-primary px-4 font-semibold text-on-accent"
			>{text.respond}</a
		>
	{/if}
	{#if !card.is_own}
		<LazyReportListing target={{ target_type: 'NEED', target_id: card.id }} name={card.title} />
	{/if}
</article>
