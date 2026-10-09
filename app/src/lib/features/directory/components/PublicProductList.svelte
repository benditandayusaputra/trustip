<script lang="ts">
	import Package from '@lucide/svelte/icons/package';
	import type { PublicConsignor } from '#lib/api/public.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		products: PublicConsignor['products'];
	}

	let { products }: Props = $props();
</script>

{#if products.length === 0}
	<p class="text-base text-text-muted">{id.publicProfile.noProducts}</p>
{:else}
	<ul class="grid grid-cols-2 gap-4 md:grid-cols-4">
		{#each products as product (product.sku)}
			<li class="flex flex-col gap-2">
				{#if product.photo_url}
					<img
						src={product.photo_url}
						alt=""
						width="200"
						height="200"
						loading="lazy"
						decoding="async"
						class="aspect-square w-full rounded-control object-cover"
					/>
				{:else}
					<Package class="aspect-square size-full text-text-muted" aria-hidden="true" />
				{/if}
				<span class="text-base text-text">{product.name}</span>
				<span class="text-sm text-text-muted">{product.category}</span>
				<span class="font-mono text-base text-text"
					>{formatRupiah(product.suggested_retail_price)}</span
				>
			</li>
		{/each}
	</ul>
{/if}
