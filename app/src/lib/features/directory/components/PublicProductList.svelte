<script lang="ts">
	import Package from '@lucide/svelte/icons/package';
	import type { PublicConsignor } from '#lib/api/public.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { publicProfile as id } from '#lib/i18n/public-profile.js';

	interface Props {
		products: PublicConsignor['products'];
		consignorSlug: string;
	}

	let { products, consignorSlug }: Props = $props();
</script>

{#if products.length === 0}
	<p class="text-base text-text-muted">{id.publicProfile.noProducts}</p>
{:else}
	<ul class="grid grid-cols-(--product-grid-columns) gap-5">
		{#each products as product (product.sku)}
			<li class="product-tile">
				{#if product.photo_url}
					<img
						src={product.photo_url}
						alt=""
						width="200"
						height="200"
						loading="lazy"
						decoding="async"
						class="product-photo"
					/>
				{:else}
					<span class="product-photo"><Package class="size-12" aria-hidden="true" /></span>
				{/if}
				<a
					href="/p/{encodeURIComponent(consignorSlug)}/{encodeURIComponent(product.sku)}"
					class="text-base font-semibold text-accent">{product.name}</a
				>
				<span class="text-sm text-text-muted">{product.category}</span>
				<span class="font-mono text-base text-text"
					>{formatRupiah(product.suggested_retail_price)}</span
				>
			</li>
		{/each}
	</ul>
{/if}
