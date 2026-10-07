<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Product } from '#lib/api/products.js';
	import ProductForm from '#lib/features/products/components/ProductForm.svelte';
	import { id } from '#lib/i18n/id.js';
</script>

<svelte:head>
	<title>{id.products.newTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.products.newTitle}</h1>
	<ProductForm product={null} onsaved={(saved: Product) => void goto(`/app/produk/${saved.id}`)} />
</section>
