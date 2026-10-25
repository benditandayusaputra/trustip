<script lang="ts">
	import Megaphone from '@lucide/svelte/icons/megaphone';
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { NeedBoardCard } from '#lib/api/directory.js';
	import { formatBasisPoints, formatDate, formatDistance, formatRupiah } from '#lib/i18n/format.js';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { organization as organizationText } from '#lib/i18n/organization.js';

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

<article class="surface-card list-card">
	{#if card.store.logo_url}
		<img
			src={card.store.logo_url}
			alt=""
			width="400"
			height="225"
			loading="lazy"
			decoding="async"
			class="aspect-video w-full rounded-control object-cover"
		/>
	{:else}
		<div class="flex aspect-video items-center justify-center rounded-control bg-accent-surface">
			<Megaphone class="size-12 text-accent" aria-hidden="true" />
		</div>
	{/if}
	<header class="flex items-start gap-4">
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<h2 class="line-clamp-2 text-lg font-semibold text-text" title={card.title}>
				<a href="/kebutuhan/{card.id}" class="list-card-link hover:text-accent">{card.title}</a>
			</h2>
			<p class="truncate text-sm text-text-muted">
				{card.store.name}
				{#if card.store.store_type}
					· {id.storeProfile.storeTypes[card.store.store_type]}
				{/if}
				{#if card.distance_m !== null}
					· <span class="font-mono">{formatDistance(card.distance_m)}</span>
				{/if}
				· {card.store.city_name}
			</p>
		</div>
	</header>
	<div class="flex flex-wrap gap-2">
		<Badge tone="neutral" label={card.category.name} />
		{#if card.is_own}
			<Badge tone="info" label={text.own} />
		{/if}
	</div>
	{#if card.description}
		<p class="line-clamp-2 text-base text-text" title={card.description}>{card.description}</p>
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
		<div class="list-card-actions">
			<a
				href={respondHref}
				class="inline-flex min-h-11 items-center rounded-control bg-primary px-4 font-semibold text-on-accent"
				>{text.respond}</a
			>
		</div>
	{/if}
</article>
