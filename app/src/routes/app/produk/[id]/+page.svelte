<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import PackageOpen from '@lucide/svelte/icons/package-open';
	import { page } from '$app/state';
	import ProductDetail from '#lib/features/products/components/ProductDetail.svelte';
	import { products as id } from '#lib/i18n/products.js';
</script>

<svelte:head>
	<title>{id.products.editTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<PageHeading icon={PackageOpen}>{id.products.editTitle}</PageHeading>
	<ProductDetail productId={page.params.id ?? ''} />
</section>
