<script lang="ts">
	import Plus from '@lucide/svelte/icons/plus';
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import PackageOpen from '@lucide/svelte/icons/package-open';
	import ProductCatalog from '#lib/features/products/components/ProductCatalog.svelte';
	import { products as id } from '#lib/i18n/products.js';
</script>

<svelte:head>
	<title>{id.products.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<PageHeading icon={PackageOpen}>{id.products.title}</PageHeading>
		<a href="/app/produk/baru" class="action-link">
			<Plus class="size-5" aria-hidden="true" />
			{id.products.add}
		</a>
	</div>
	<ProductCatalog />
</section>
