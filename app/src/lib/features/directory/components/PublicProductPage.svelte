<script lang="ts">
	import ChefHat from '@lucide/svelte/icons/chef-hat';
	import Package from '@lucide/svelte/icons/package';
	import Card from '#lib/components/ui/Card.svelte';
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import ShareButton from '#lib/components/ui/ShareButton.svelte';
	import type { PublicProduct } from '#lib/api/public.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { publicProfile as id } from '#lib/i18n/public-profile.js';
	import type { GeoPosition } from '../availability.svelte';
	import AvailableStores from './AvailableStores.svelte';
	import DeclaredLegality from './DeclaredLegality.svelte';
	import LazyReportListing from './LazyReportListing.svelte';

	interface Props {
		product: PublicProduct;
		consignorSlug: string;
		locate: () => Promise<GeoPosition>;
		shareUrl: string;
	}

	let { product, consignorSlug, locate, shareUrl }: Props = $props();

	const text = id.publicProfile;
	const maker = $derived(product.maker);
</script>

<article class="flex flex-col gap-6">
	<PageHeading icon={Package} backHref="/cari/penitip">{product.name}</PageHeading>
	<header class="grid items-center gap-6 lg:grid-cols-2">
		{#if product.photo_url}
			<img
				src={product.photo_url}
				alt={product.name}
				width="400"
				height="300"
				decoding="async"
				class="aspect-4/3 w-full rounded-card border border-border object-cover"
			/>
		{:else}
			<div
				class="flex aspect-4/3 items-center justify-center rounded-card border border-border bg-surface-hover"
			>
				<Package class="size-12 text-text-muted" aria-hidden="true" />
			</div>
		{/if}
		<div class="flex flex-col gap-4">
			<p class="section-label">{product.category}</p>
			<a
				href="/u/{consignorSlug}"
				class="inline-flex min-h-11 items-center gap-2 self-start text-base font-semibold text-accent hover:underline"
				><ChefHat class="size-5" aria-hidden="true" />{text.productsBy(maker.name)}</a
			>
			<p class="font-mono text-kpi-compact font-bold text-text">
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
	<ShareButton title={product.name} url={shareUrl} />
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
	<LazyReportListing
		target={{ target_type: 'PRODUCT', target_id: product.id }}
		name={product.name}
	/>
</article>
