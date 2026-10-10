<script lang="ts">
	import { page } from '$app/state';
	import ProductDetail from '#lib/features/products/components/ProductDetail.svelte';
	import { products as id } from '#lib/i18n/products.js';
</script>

<svelte:head>
	<title>{id.products.editTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.products.editTitle}</h1>
	<ProductDetail productId={page.params.id ?? ''} />
</section>
