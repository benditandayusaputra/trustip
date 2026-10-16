<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import PackageOpen from '@lucide/svelte/icons/package-open';
	import { goto } from '$app/navigation';
	import type { Product } from '#lib/api/products.js';
	import ProductForm from '#lib/features/products/components/ProductForm.svelte';
	import { products as id } from '#lib/i18n/products.js';
</script>

<svelte:head>
	<title>{id.products.newTitle}</title>
</svelte:head>

<section class="trade-panel mx-auto flex w-full max-w-form flex-col gap-6">
	<PageHeading icon={PackageOpen}>{id.products.newTitle}</PageHeading>
	<ProductForm product={null} onsaved={(saved: Product) => void goto(`/app/produk/${saved.id}`)} />
</section>
