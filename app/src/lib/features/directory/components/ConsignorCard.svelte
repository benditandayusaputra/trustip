<script lang="ts">
	import ChefHat from '@lucide/svelte/icons/chef-hat';
	import MapPin from '@lucide/svelte/icons/map-pin';
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { DirectoryConsignorCard } from '#lib/api/directory.js';
	import { formatBusinessDate, formatDistance } from '#lib/i18n/format.js';
	import { directory as id } from '#lib/i18n/directory.js';
	import ConsignorTraits from './ConsignorTraits.svelte';
	import ContactActions from './ContactActions.svelte';
	import MatchReasons from './MatchReasons.svelte';
	import MatchSummary from './MatchSummary.svelte';

	interface Props {
		card: DirectoryConsignorCard;
	}

	let { card }: Props = $props();

	const text = id.consignorDirectory;
	const shownCategories = 2;
	const hiddenCategories = $derived(card.categories.length - shownCategories);
</script>

<article class="surface-card list-card">
	<div class="relative">
		{#if card.logo_url}
			<img
				src={card.logo_url}
				alt=""
				width="400"
				height="300"
				loading="lazy"
				decoding="async"
				class="aspect-4/3 w-full rounded-control object-cover"
			/>
		{:else}
			<div class="flex aspect-4/3 items-center justify-center rounded-control bg-accent-surface">
				<ChefHat class="size-12 text-accent" aria-hidden="true" />
			</div>
		{/if}
		{#if card.distance_m !== null}
			<span
				class="absolute top-2 left-2 rounded-pill bg-surface px-2 py-1 font-mono text-xs font-semibold text-text shadow-card"
				>{text.approximate(formatDistance(card.distance_m))}</span
			>
		{/if}
	</div>
	<header class="flex min-w-0 flex-col gap-1">
		<h2 class="line-clamp-2 text-lg font-semibold text-text" title={card.name}>
			<a href="/u/{card.slug}" class="list-card-link hover:text-accent">{card.name}</a>
		</h2>
		<p class="flex min-w-0 items-center gap-1 text-sm text-text-muted">
			<MapPin class="size-4 shrink-0" aria-hidden="true" /><span class="truncate"
				>{card.district ? `${card.district}, ` : ''}{card.city_name}</span
			>
		</p>
		<p class="text-sm font-semibold text-text">{text.productCount(card.product_count)}</p>
		<p class="text-xs text-text-muted">{id.directory.joined(formatBusinessDate(card.joined_on))}</p>
	</header>
	<div class="flex flex-wrap gap-2">
		{#if card.is_self}
			<Badge tone="neutral" label={id.directory.self} />
		{/if}
		{#if card.is_partner}
			<Badge tone="success" label={id.directory.partner} />
		{/if}
		{#each card.categories.slice(0, shownCategories) as category (category.id)}
			<Badge tone="neutral" label={category.name} />
		{/each}
		{#if hiddenCategories > 0}
			<Badge tone="neutral" label={id.directory.more(hiddenCategories)} />
		{/if}
	</div>
	<ConsignorTraits traits={card.traits} />
	{#if card.match}
		<MatchSummary match={card.match} compact />
		<MatchReasons reasons={card.match.reasons} isApproximate />
	{/if}
	{#if !card.is_self}
		<div class="list-card-actions">
			<ContactActions
				organizationId={card.id}
				name={card.name}
				canPropose={!card.is_partner}
				stacked
			/>
		</div>
	{/if}
</article>
