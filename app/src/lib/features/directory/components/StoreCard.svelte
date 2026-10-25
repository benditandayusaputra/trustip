<script lang="ts">
	import Store from '@lucide/svelte/icons/store';
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { DirectoryStoreCard } from '#lib/api/directory.js';
	import { formatBasisPoints, formatBusinessDate, formatDistance } from '#lib/i18n/format.js';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { organization as organizationText } from '#lib/i18n/organization.js';
	import ContactActions from './ContactActions.svelte';
	import MatchReasons from './MatchReasons.svelte';
	import MatchSummary from './MatchSummary.svelte';

	const id = { ...directoryText, ...organizationText };

	interface Props {
		card: DirectoryStoreCard;
	}

	let { card }: Props = $props();

	const text = id.directory;
	const shownCategories = 2;
	const hiddenCategories = $derived(card.accepted_categories.length - shownCategories);
	const storeTypes = id.storeProfile.storeTypes;
	const minimum = $derived(card.terms.commission_min_bps);
	const maximum = $derived(card.terms.commission_max_bps);
	const reputation = $derived(
		card.reputation.on_time_bps === null
			? text.newcomer
			: text.onTime(formatBasisPoints(card.reputation.on_time_bps))
	);
</script>

<article class="surface-card list-card">
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
			<div
				class="flex size-12 shrink-0 items-center justify-center rounded-control bg-accent-surface"
			>
				<Store class="size-6 text-accent" aria-hidden="true" />
			</div>
		{/if}
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<h2 class="truncate text-lg font-semibold text-text" title={card.name}>
				<a href="/t/{card.slug}" class="list-card-link hover:text-accent">{card.name}</a>
			</h2>
			<p class="truncate text-sm text-text-muted">
				{card.store_type ? storeTypes[card.store_type] : ''}
				{#if card.distance_m !== null}
					· <span class="font-mono">{formatDistance(card.distance_m)}</span>
				{/if}
				{card.store_type || card.distance_m !== null ? '·' : ''}
				{card.city_name}
			</p>
		</div>
	</header>
	{#if card.match}<MatchSummary match={card.match} compact />{/if}
	<div class="flex flex-wrap gap-2">
		{#if card.is_self}
			<Badge tone="neutral" label={text.self} />
		{/if}
		{#if card.is_partner}
			<Badge tone="success" label={text.partner} />
		{/if}
		{#if card.open_need_category_ids.length > 0}
			<Badge tone="pending" label={text.openNeed} />
		{/if}
		{#each card.accepted_categories.slice(0, shownCategories) as category (category.id)}
			<Badge tone="neutral" label={category.name} />
		{/each}
		{#if hiddenCategories > 0}
			<Badge tone="neutral" label={text.more(hiddenCategories)} />
		{/if}
	</div>
	<div class="flex flex-col gap-1 text-sm text-text-muted">
		{#if minimum !== null && maximum !== null}
			<p class="truncate">
				{text.commissionRange(formatBasisPoints(minimum), formatBasisPoints(maximum))}
			</p>
		{/if}
		<p class="truncate">{reputation} · {text.activePartners(card.reputation.active_partners)}</p>
		<p class="truncate">{text.joined(formatBusinessDate(card.joined_on))}</p>
	</div>
	{#if card.match}
		<MatchReasons reasons={card.match.reasons} />
	{/if}
	{#if !card.is_self}
		<div class="list-card-actions">
			<ContactActions organizationId={card.id} name={card.name} canPropose={!card.is_partner} />
		</div>
	{/if}
</article>
