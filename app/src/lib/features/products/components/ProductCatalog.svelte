<script lang="ts">
	import PackageOpen from '@lucide/svelte/icons/package-open';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatRupiah } from '#lib/i18n/format.js';
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
	<div class="grid gap-4 md:grid-cols-3">
		<TextField label={text.search} type="search" name="q" bind:value={catalog.search.text} />
		<Select
			label={text.category}
			options={categoryOptions}
			bind:value={catalog.search.categoryId}
		/>
		<Select label={text.status} options={statusOptions} bind:value={catalog.search.status} />
	</div>
	{#if catalog.status === 'loading'}
		<Skeleton />
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
		<ul class="flex flex-col gap-2">
			{#each catalog.items as product (product.id)}
				<li>
					<a
						href="/app/produk/{product.id}"
						class="flex items-center justify-between gap-4 rounded-control border border-border bg-surface p-4 hover:bg-surface-hover"
					>
						<span class="flex flex-col gap-1">
							<span class="text-base font-semibold text-text">{product.name}</span>
							<span class="text-sm text-text-muted">{product.sku} · {product.category_name}</span>
						</span>
						<span class="flex flex-col items-end gap-1">
							<span class="font-mono font-medium text-text tabular-nums">
								{formatRupiah(product.suggested_retail_price)}
							</span>
							{#if product.status === 'ARCHIVED'}
								<Badge tone="neutral" label={text.statuses.ARCHIVED} />
							{/if}
						</span>
					</a>
				</li>
			{/each}
		</ul>
		{#if catalog.nextCursor}
			<Button variant="secondary" busy={catalog.busy} onclick={() => void catalog.loadMore()}>
				{text.loadMore}
			</Button>
		{/if}
	{/if}
</div>
