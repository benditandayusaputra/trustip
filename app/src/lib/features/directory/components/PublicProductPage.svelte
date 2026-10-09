<script lang="ts">
	import Package from '@lucide/svelte/icons/package';
	import Card from '#lib/components/ui/Card.svelte';
	import type { PublicProduct } from '#lib/api/public.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import type { GeoPosition } from '../availability.svelte';
	import AvailableStores from './AvailableStores.svelte';
	import DeclaredLegality from './DeclaredLegality.svelte';
	import ReportListing from './ReportListing.svelte';

	interface Props {
		product: PublicProduct;
		consignorSlug: string;
		locate: () => Promise<GeoPosition>;
	}

	let { product, consignorSlug, locate }: Props = $props();

	const text = id.publicProfile;
	const maker = $derived(product.maker);
</script>

<article class="flex flex-col gap-6">
	<header class="flex flex-col gap-4 md:flex-row">
		{#if product.photo_url}
			<img
				src={product.photo_url}
				alt={product.name}
				width="400"
				height="300"
				decoding="async"
				class="aspect-4/3 w-full rounded-control object-cover md:max-w-form"
			/>
		{:else}
			<Package class="size-12 text-text-muted" aria-hidden="true" />
		{/if}
		<div class="flex flex-col gap-2">
			<p class="text-sm text-text-muted">{product.category}</p>
			<h1 class="text-xl font-semibold text-text">{product.name}</h1>
			<p class="font-mono text-kpi font-bold text-accent">
				{formatRupiah(product.suggested_retail_price)}
			</p>
			<p class="text-sm text-text-muted">{text.suggestedPrice(product.unit)}</p>
		</div>
	</header>
	{#if product.description}
		<p class="text-base whitespace-pre-line text-text">{product.description}</p>
	{/if}
	{#if product.composition || product.net_content}
		<Card title={text.composition}>
			{#if product.composition}
				<p class="text-base whitespace-pre-line text-text">{product.composition}</p>
			{/if}
			{#if product.net_content}
				<p class="text-sm text-text-muted">{text.netContent(product.net_content)}</p>
			{/if}
		</Card>
	{/if}
	<Card title={text.availableAt}>
		<AvailableStores {product} {consignorSlug} {locate} />
	</Card>
	<Card title={text.maker}>
		{#if maker.profile_slug}
			<a href="/u/{maker.profile_slug}" class="text-base font-semibold text-accent">{maker.name}</a>
		{:else}
			<p class="text-base font-semibold text-text">{maker.name}</p>
		{/if}
		<p class="text-sm text-text-muted">{text.place(maker.district, maker.city_name)}</p>
		<h3 class="text-base font-semibold text-text">{text.legality}</h3>
		<DeclaredLegality legality={maker.legality} />
	</Card>
	<ReportListing target={{ target_type: 'PRODUCT', target_id: product.id }} name={product.name} />
</article>
