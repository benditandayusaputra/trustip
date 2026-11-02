<script lang="ts">
	import MapPin from '@lucide/svelte/icons/map-pin';
	import Badge from '#lib/components/ui/Badge.svelte';
	import StoreImage from '#lib/components/ui/StoreImage.svelte';
	import type { DirectoryStoreCard } from '#lib/api/directory.js';
	import { formatBasisPoints, formatBusinessDate, formatDistance } from '#lib/i18n/format.js';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { organization as organizationText } from '#lib/i18n/organization.js';
	import { publicProfile as publicProfileText } from '#lib/i18n/public-profile.js';
	import ContactActions from './ContactActions.svelte';
	import MatchReasons from './MatchReasons.svelte';
	import MatchSummary from './MatchSummary.svelte';
	import StoreFacts from './StoreFacts.svelte';

	const id = { ...directoryText, ...organizationText, ...publicProfileText };

	interface Props {
		card: DirectoryStoreCard;
	}

	let { card }: Props = $props();

	const text = id.directory;
	const shownCategories = 2;
	const hiddenCategories = $derived(card.accepted_categories.length - shownCategories);
	const minimum = $derived(card.terms.commission_min_bps);
	const maximum = $derived(card.terms.commission_max_bps);
	const facts = $derived([
		...(minimum !== null && maximum !== null
			? [
					{
						label: text.commissionLabel,
						value: text.percentRange(formatBasisPoints(minimum), formatBasisPoints(maximum))
					}
				]
			: []),
		{
			label: text.onTimeLabel,
			value:
				card.reputation.on_time_bps === null
					? text.newcomer
					: formatBasisPoints(card.reputation.on_time_bps)
		},
		{ label: text.partnersLabel, value: String(card.reputation.active_partners) }
	]);
</script>

<article class="surface-card list-card">
	<div class="relative">
		<StoreImage url={card.logo_url} class="aspect-video w-full rounded-control bg-accent-surface" />
		{#if card.distance_m !== null}
			<span
				class="absolute top-2 left-2 rounded-pill bg-surface px-2 py-1 font-mono text-xs font-semibold text-text shadow-card"
				>{formatDistance(card.distance_m)}</span
			>
		{/if}
	</div>
	<header class="flex min-w-0 flex-col gap-1">
		{#if card.store_type}
			<p class="section-label truncate">{id.storeProfile.storeTypes[card.store_type]}</p>
		{/if}
		<h2 class="truncate text-lg font-semibold text-text" title={card.name}>
			<a href="/t/{card.slug}" class="list-card-link hover:text-accent">{card.name}</a>
		</h2>
		<p class="flex min-w-0 items-center gap-1 text-sm text-text-muted">
			<MapPin class="size-4 shrink-0" aria-hidden="true" /><span class="truncate"
				>{id.publicProfile.place(card.district, card.city_name)}</span
			>
		</p>
		<p class="text-xs text-text-muted">{text.joined(formatBusinessDate(card.joined_on))}</p>
	</header>
	<StoreFacts {facts} />
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
	{#if card.match}
		<MatchSummary match={card.match} compact />
		<MatchReasons reasons={card.match.reasons} />
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
