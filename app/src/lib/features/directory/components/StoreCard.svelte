<script lang="ts">
	import Store from '@lucide/svelte/icons/store';
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { DirectoryStoreCard } from '#lib/api/directory.js';
	import { formatBasisPoints, formatBusinessDate, formatDistance } from '#lib/i18n/format.js';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { organization as organizationText } from '#lib/i18n/organization.js';
	import ContactActions from './ContactActions.svelte';
	import LazyReportListing from './LazyReportListing.svelte';
	import MatchReasons from './MatchReasons.svelte';
	import MatchSummary from './MatchSummary.svelte';

	const id = { ...directoryText, ...organizationText };

	interface Props {
		card: DirectoryStoreCard;
	}

	let { card }: Props = $props();

	const text = id.directory;
	const storeTypes = id.storeProfile.storeTypes;
	const minimum = $derived(card.terms.commission_min_bps);
	const maximum = $derived(card.terms.commission_max_bps);
	const reputation = $derived(
		card.reputation.on_time_bps === null
			? text.newcomer
			: text.onTime(formatBasisPoints(card.reputation.on_time_bps))
	);
</script>

<article class="flex min-w-0 flex-col gap-4 rounded-card border border-border bg-surface p-5">
	<header class="flex flex-wrap items-start gap-4">
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
			<h2 class="text-lg font-semibold text-text">
				<a href="/t/{card.slug}" class="break-words hover:text-accent">{card.name}</a>
			</h2>
			<p class="text-sm text-text-muted">
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
		{#each card.accepted_categories as category (category.id)}
			<Badge tone="neutral" label={category.name} />
		{/each}
	</div>
	<div class="flex flex-col gap-1 text-sm text-text-muted">
		{#if minimum !== null && maximum !== null}
			<p>{text.commissionRange(formatBasisPoints(minimum), formatBasisPoints(maximum))}</p>
		{/if}
		<p>{reputation} · {text.activePartners(card.reputation.active_partners)}</p>
		<p>{text.joined(formatBusinessDate(card.joined_on))}</p>
	</div>
	{#if card.match}
		<MatchReasons reasons={card.match.reasons} />
	{/if}
	{#if !card.is_self}
		<ContactActions organizationId={card.id} name={card.name} canPropose={!card.is_partner} />
		<LazyReportListing
			target={{ target_type: 'ORGANIZATION', target_id: card.id }}
			name={card.name}
		/>
	{/if}
</article>
