<script lang="ts">
	import PackageOpen from '@lucide/svelte/icons/package-open';
	import Alert from '#lib/components/ui/Alert.svelte';
	import CatalogProductCard from './CatalogProductCard.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { common as commonText } from '#lib/i18n/common.js';
	import { products as productsText } from '#lib/i18n/products.js';
	import { ProductCatalog } from '../catalog.svelte';

	const id = { ...commonText, ...productsText };

	const text = id.products;
	const catalog = new ProductCatalog(apiClient);
	const statusOptions = Object.entries(text.statuses).map(([value, label]) => ({ value, label }));
	const categoryOptions = $derived([
		{ value: '', label: text.allCategories },
		...catalog.categories.map((category) => ({ value: category.id, label: category.name }))
	]);
	const isFiltered = $derived(catalog.search.text !== '' || catalog.search.categoryId !== '');

	$effect(() => {
		catalog.searchSoon();
	});

	$effect(() => {
		void catalog.loadCategories();
	});
</script>

<div class="flex flex-col gap-4">
	<div class="filter-toolbar grid gap-4 md:grid-cols-3">
		<TextField label={text.search} type="search" name="q" bind:value={catalog.search.text} />
		<Select
			label={text.category}
			options={categoryOptions}
			bind:value={catalog.search.categoryId}
		/>
		<Select label={text.status} options={statusOptions} bind:value={catalog.search.status} />
	</div>
	{#if catalog.status === 'loading'}
		<ContentSkeleton variant="catalog" />
	{:else if catalog.status === 'failed'}
		<Alert title={text.loadFailed}>
			{catalog.failure}
			{#snippet action()}
				<Button variant="secondary" onclick={() => void catalog.load()}>{id.auth.retry}</Button>
			{/snippet}
		</Alert>
	{:else if catalog.items.length === 0}
		<EmptyState icon={PackageOpen} message={isFiltered ? text.noMatch : text.empty}>
			{#snippet action()}
				<a href="/app/produk/baru" class="font-semibold text-accent">{text.add}</a>
			{/snippet}
		</EmptyState>
	{:else}
		<ul class="grid auto-rows-fr grid-cols-(--product-grid-columns) gap-5">
			{#each catalog.items as product (product.id)}
				<li class="min-w-0"><CatalogProductCard {product} /></li>
			{/each}
		</ul>
		{#if catalog.nextCursor}
			<Button variant="secondary" busy={catalog.busy} onclick={() => void catalog.loadMore()}>
				{text.loadMore}
			</Button>
		{/if}
	{/if}
</div>
