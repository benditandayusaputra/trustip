<script lang="ts">
	import ChefHat from '@lucide/svelte/icons/chef-hat';
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { DirectoryConsignorCard } from '#lib/api/directory.js';
	import { formatBusinessDate, formatDistance } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import ConsignorTraits from './ConsignorTraits.svelte';
	import ContactActions from './ContactActions.svelte';
	import ReportListing from './ReportListing.svelte';
	import MatchReasons from './MatchReasons.svelte';
	import MatchSummary from './MatchSummary.svelte';

	interface Props {
		card: DirectoryConsignorCard;
	}

	let { card }: Props = $props();

	const text = id.consignorDirectory;
</script>

<article class="flex flex-col gap-4 rounded-card border border-border bg-surface p-4">
	<header class="flex items-start gap-4">
		{#if card.logo_url}
			<img
				src={card.logo_url}
				alt=""
				width="48"
				height="48"
				loading="lazy"
				decoding="async"
				class="size-12 shrink-0 rounded-control object-cover"
			/>
		{:else}
			<ChefHat class="size-12 shrink-0 text-text-muted" aria-hidden="true" />
		{/if}
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<h2 class="text-lg font-semibold text-text">{card.name}</h2>
			<p class="text-sm text-text-muted">
				{text.productCount(card.product_count)}
				{#if card.distance_m !== null}
					· <span class="font-mono">{text.approximate(formatDistance(card.distance_m))}</span>
				{/if}
				· {card.district ? `${card.district}, ` : ''}{card.city_name}
			</p>
			<p class="text-sm text-text-muted">
				{id.directory.joined(formatBusinessDate(card.joined_on))}
			</p>
		</div>
		{#if card.match}
			<MatchSummary match={card.match} />
		{/if}
	</header>
	<div class="flex flex-wrap gap-2">
		{#if card.is_self}
			<Badge tone="neutral" label={id.directory.self} />
		{/if}
		{#if card.is_partner}
			<Badge tone="success" label={id.directory.partner} />
		{/if}
		{#each card.categories as category (category.id)}
			<Badge tone="neutral" label={category.name} />
		{/each}
	</div>
	<ConsignorTraits traits={card.traits} />
	{#if card.match}
		<MatchReasons reasons={card.match.reasons} isApproximate />
	{/if}
	{#if !card.is_self}
		<ContactActions organizationId={card.id} name={card.name} canPropose={!card.is_partner} />
		<ReportListing target={{ target_type: 'ORGANIZATION', target_id: card.id }} name={card.name} />
	{/if}
</article>
