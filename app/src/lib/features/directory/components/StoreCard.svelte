<script lang="ts">
	import Store from '@lucide/svelte/icons/store';
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { DirectoryStoreCard } from '#lib/api/directory.js';
	import { formatBasisPoints, formatDistance } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { reasonText } from '../reasons';

	interface Props {
		card: DirectoryStoreCard;
	}

	let { card }: Props = $props();

	const text = id.directory;
	const storeTypes = id.storeProfile.storeTypes;
	const minimum = $derived(card.terms.commission_min_bps);
	const maximum = $derived(card.terms.commission_max_bps);
	const label = $derived(card.match?.label ? text.labels[card.match.label] : undefined);
	const reputation = $derived(
		card.reputation.on_time_bps === null
			? text.newcomer
			: text.onTime(formatBasisPoints(card.reputation.on_time_bps))
	);
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
			<Store class="size-12 shrink-0 text-text-muted" aria-hidden="true" />
		{/if}
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<h2 class="text-lg font-semibold text-text">{card.name}</h2>
			<p class="text-sm text-text-muted">
				{card.store_type ? storeTypes[card.store_type] : ''}
				{#if card.distance_m !== null}
					· <span class="font-mono">{formatDistance(card.distance_m)}</span>
				{/if}
				· {card.city_name}
			</p>
		</div>
		{#if card.match}
			<p class="flex shrink-0 flex-col items-end gap-1">
				<span
					class="text-2xl font-mono font-bold text-accent"
					aria-label={text.score(card.match.score)}>{card.match.score}</span
				>
				{#if label}
					<Badge tone="info" {label} />
				{/if}
			</p>
		{/if}
	</header>
	<div class="flex flex-wrap gap-2">
		{#if card.is_self}
			<Badge tone="neutral" label={text.self} />
		{/if}
		{#if card.is_partner}
			<Badge tone="success" label={text.partner} />
		{/if}
		{#each card.accepted_categories as category (category.id)}
			<Badge tone="neutral" label={category.name} />
		{/each}
	</div>
	<div class="flex flex-col gap-1 text-sm text-text-muted">
		{#if minimum !== null && maximum !== null}
			<p>{text.commissionRange(formatBasisPoints(minimum), formatBasisPoints(maximum))}</p>
		{/if}
		<p>{reputation}</p>
	</div>
	{#if card.match && card.match.reasons.length > 0}
		<ul class="flex flex-wrap gap-2 text-sm text-text">
			{#each card.match.reasons as reason (reason.code)}
				<li class="rounded-pill bg-surface-hover px-2 py-1">{reasonText(reason)}</li>
			{/each}
		</ul>
	{/if}
</article>
