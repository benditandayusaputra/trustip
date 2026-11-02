<script lang="ts">
	import PackageOpen from '@lucide/svelte/icons/package-open';
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { DirectoryProductCard } from '#lib/api/directory.js';
	import { formatDistance, formatRupiah } from '#lib/i18n/format.js';
	import { directory as id } from '#lib/i18n/directory.js';
	import ConsignorTraits from './ConsignorTraits.svelte';
	import ContactActions from './ContactActions.svelte';
	import MatchReasons from './MatchReasons.svelte';
	import MatchSummary from './MatchSummary.svelte';

	interface Props {
		card: DirectoryProductCard;
	}

	let { card }: Props = $props();

	const text = id.consignorDirectory;
</script>

<article class="surface-card list-card">
	{#if card.photo_url}
		<img
			src={card.photo_url}
			alt=""
			width="400"
			height="300"
			loading="lazy"
			decoding="async"
			class="aspect-4/3 w-full rounded-control object-cover"
		/>
	{:else}
		<div class="flex aspect-4/3 items-center justify-center rounded-control bg-surface-hover">
			<PackageOpen class="size-12 text-text-muted" aria-hidden="true" />
		</div>
	{/if}
	<header class="flex items-start gap-4">
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<h2 class="line-clamp-2 text-lg font-semibold text-text" title={card.name}>
				<a href="/p/{card.consignor.slug}/{card.sku}" class="list-card-link hover:text-accent"
					>{card.name}</a
				>
			</h2>
			<p class="truncate text-sm text-text-muted">
				{text.madeBy(card.consignor.name)}
				{#if card.distance_m !== null}
					· <span class="font-mono">{text.approximate(formatDistance(card.distance_m))}</span>
				{/if}
			</p>
			<p class="font-mono text-lg font-semibold text-text">
				{formatRupiah(card.suggested_retail_price)}
			</p>
		</div>
		{#if card.match}
			<MatchSummary match={card.match} />
		{/if}
	</header>
	<div class="flex flex-wrap gap-2">
		{#if card.is_partner}
			<Badge tone="success" label={id.directory.partner} />
		{/if}
		<Badge tone="neutral" label={card.category.name} />
		{#if card.shelf_life_days !== null}
			<Badge tone="neutral" label={text.shelfLifeDays(card.shelf_life_days)} />
		{/if}
	</div>
	<ConsignorTraits traits={card.traits} />
	{#if card.match}
		<MatchReasons reasons={card.match.reasons} isApproximate />
	{/if}
	{#if !card.is_self}
		<div class="list-card-actions">
			<ContactActions
				organizationId={card.consignor.id}
				name={card.consignor.name}
				productId={card.id}
				canPropose={!card.is_partner}
				stacked
			/>
		</div>
	{/if}
</article>
