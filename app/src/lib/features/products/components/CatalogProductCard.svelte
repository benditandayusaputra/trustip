<script lang="ts">
	import PackageOpen from '@lucide/svelte/icons/package-open';
	import ArrowUpRight from '@lucide/svelte/icons/arrow-up-right';
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { Product } from '#lib/api/products.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { products as id } from '#lib/i18n/products.js';
	let { product }: { product: Product } = $props();
</script>

<a href="/app/produk/{product.id}" class="product-tile">
	{#if product.photo}
		<img
			src={product.photo.url}
			alt=""
			width="400"
			height="300"
			loading="lazy"
			decoding="async"
			class="product-photo"
		/>
	{:else}
		<span class="product-photo"><PackageOpen class="size-12" aria-hidden="true" /></span>
	{/if}
	<span class="flex min-w-0 flex-col gap-2">
		<span class="flex flex-wrap items-center justify-between gap-2">
			<span class="text-xs text-text-muted">{product.category_name}</span>
			<Badge
				tone={product.status === 'ACTIVE' ? 'success' : 'neutral'}
				label={id.products.statuses[product.status]}
			/>
		</span>
		<span class="text-lg font-semibold wrap-anywhere text-text">{product.name}</span>
		<span class="font-mono text-xs text-text-muted">{product.sku}</span>
	</span>
	<span class="mt-auto flex items-center justify-between gap-2 border-t border-border pt-4">
		<span class="font-mono text-base font-semibold wrap-anywhere text-text tabular-nums"
			>{formatRupiah(product.suggested_retail_price)}</span
		>
		<ArrowUpRight class="size-5 shrink-0 text-accent" aria-hidden="true" />
	</span>
</a>
