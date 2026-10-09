<script lang="ts">
	import Badge from '#lib/components/ui/Badge.svelte';
	import Card from '#lib/components/ui/Card.svelte';
	import ShareButton from '#lib/components/ui/ShareButton.svelte';
	import type { PublicNeed } from '#lib/api/public.js';
	import { formatBasisPoints, formatDate, formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import ReportListing from './ReportListing.svelte';

	interface Props {
		need: PublicNeed;
		shareUrl: string;
	}

	let { need, shareUrl }: Props = $props();

	const text = id.publicProfile;
	const board = id.needBoard;
	const isOpen = $derived(need.status === 'OPEN');
	const respondHref = $derived(
		`/app/ajuan/baru?${new URLSearchParams({ sasaran: need.store.id, kebutuhan: need.id }).toString()}`
	);
	const store = $derived(need.store);
</script>

<article class="flex flex-col gap-6">
	<header class="flex flex-col gap-2">
		<p class="text-sm text-text-muted">{text.needKind}</p>
		<h1 class="text-xl font-semibold text-text">{need.title}</h1>
		<p class="text-sm text-text-muted">
			{text.needBy(store.name)}
			{#if store.store_type}
				· {id.storeProfile.storeTypes[store.store_type]}
			{/if}
			· {text.place(store.district, store.city_name)}
		</p>
		<div class="flex flex-wrap gap-2">
			<Badge tone="neutral" label={need.category} />
			<Badge tone={isOpen ? 'info' : 'neutral'} label={id.needs.statuses[need.status]} />
		</div>
	</header>
	{#if need.description}
		<p class="text-base whitespace-pre-line text-text">{need.description}</p>
	{/if}
	<Card title={text.needDetails}>
		<ul class="flex flex-col gap-1 text-base text-text">
			{#if need.est_qty_per_week !== null}
				<li>{board.perWeek(need.est_qty_per_week)}</li>
			{/if}
			{#if need.max_retail_price !== null}
				<li>{board.priceUpTo(formatRupiah(need.max_retail_price))}</li>
			{/if}
			{#if need.preferred_model !== null}
				<li>{id.needs.models[need.preferred_model]}</li>
			{/if}
			{#if need.offered_commission_bps !== null}
				<li>{board.commission(formatBasisPoints(need.offered_commission_bps))}</li>
			{/if}
			<li>{board.expiresAt(formatDate(need.expires_at))}</li>
		</ul>
	</Card>
	<div class="flex flex-wrap items-center gap-4">
		{#if isOpen}
			<a
				href={respondHref}
				class="inline-flex min-h-11 items-center rounded-control bg-accent px-4 font-semibold text-on-accent"
				>{board.respond}</a
			>
		{:else}
			<p class="text-base text-text-muted">{text.needClosed}</p>
		{/if}
		{#if store.profile_slug}
			<a href="/t/{store.profile_slug}" class="font-semibold text-accent">{text.storeProfile}</a>
		{/if}
		<ShareButton title={need.title} url={shareUrl} />
	</div>
	<ReportListing target={{ target_type: 'NEED', target_id: need.id }} name={need.title} />
</article>
