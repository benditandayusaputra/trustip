<script lang="ts">
	import ProductCatalog from '#lib/features/products/components/ProductCatalog.svelte';
	import { products as id } from '#lib/i18n/products.js';
</script>

<svelte:head>
	<title>{id.products.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<h1 class="text-xl font-semibold text-text">{id.products.title}</h1>
		<a
			href="/app/produk/baru"
			class="inline-flex min-h-12 items-center rounded-control bg-primary px-5 font-semibold text-on-accent"
		>
			{id.products.add}
		</a>
	</div>
	<ProductCatalog />
</section>
