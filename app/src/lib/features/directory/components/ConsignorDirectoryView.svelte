<script lang="ts">
	import FilterLayout from '#lib/components/ui/FilterLayout.svelte';
	import { onMount, type Snippet } from 'svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import Tabs from '#lib/components/ui/Tabs.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { DirectoryConsignorCard, DirectoryProductCard } from '#lib/api/directory.js';
	import { listCategories, type Category } from '#lib/api/organizations.js';
	import { directory as id } from '#lib/i18n/directory.js';
	import {
		consignorSearchParams,
		defaultConsignorQuery,
		type ConsignorQuery,
		type ConsignorTab
	} from '../consignor-query';
	import { DirectoryList } from '../state.svelte';
	import { forVisitor, isRandomOrder } from '../visitor';
	import ConsignorCard from './ConsignorCard.svelte';
	import ConsignorFilters from './ConsignorFilters.svelte';
	import DirectoryResults from './DirectoryResults.svelte';
	import ProductCard from './ProductCard.svelte';

	interface Props {
		query: ConsignorQuery;
		isVisitor?: boolean;
		ready?: boolean;
		place: Snippet;
		onchange: (query: ConsignorQuery) => void;
	}

	let { query, isVisitor = false, ready = true, place, onchange }: Props = $props();

	const text = id.consignorDirectory;
	const products = new DirectoryList<DirectoryProductCard>(apiClient, '/api/v1/directory/products');
	const consignors = new DirectoryList<DirectoryConsignorCard>(
		apiClient,
		'/api/v1/directory/consignors'
	);
	const tabs = Object.entries(text.tabs).map(([key, label]) => ({ key, label }));
	let categories = $state<Category[]>([]);
	let typed = $derived(query.text);
	const typingDelayMs = 300;
	const maxSearchLength = 80;
	const search = $derived(consignorSearchParams(forVisitor(query, isVisitor)).toString());

	onMount(() => {
		listCategories(apiClient())
			.then((loaded) => (categories = loaded))
			.catch(() => (categories = []));
	});

	$effect(() => {
		if (!ready) return;
		const list = query.tab === 'produk' ? products : consignors;
		list.shuffle = isRandomOrder(query, isVisitor);
		void list.load(search);
	});

	$effect(() => {
		const next = typed;
		const timer = setTimeout(() => {
			if (next !== query.text) change({ text: next });
		}, typingDelayMs);
		return () => {
			clearTimeout(timer);
		};
	});

	function change(changes: Partial<ConsignorQuery>) {
		onchange({ ...query, ...changes });
	}

	function clear() {
		onchange({ ...defaultConsignorQuery(), tab: query.tab });
	}
</script>

<FilterLayout>
	{#snippet searchControls()}<TextField
			label={text.search}
			type="search"
			maxlength={maxSearchLength}
			bind:value={typed}
		/>{/snippet}
	{#snippet filters()}
		<ConsignorFilters {query} {categories} {isVisitor} {place} onchange={change} />
	{/snippet}

	<Tabs
		label={text.tabsLabel}
		{tabs}
		bind:selected={
			() => query.tab,
			(tab: string) => {
				change({ tab: tab as ConsignorTab });
			}
		}
	>
		{#snippet panel(selected: string)}
			{#if selected === 'produk'}
				<DirectoryResults
					list={products}
					query={search}
					skeleton="catalog"
					{isVisitor}
					failedTitle={text.productsFailed}
					emptyMessage={text.productsEmpty}
					onclear={clear}
				>
					{#snippet card(product: DirectoryProductCard)}
						<ProductCard card={product} />
					{/snippet}
					{#snippet results(items: DirectoryProductCard[])}
						<ul class="grid grid-cols-(--product-grid-columns) gap-4">
							{#each items as product (product.id)}<li><ProductCard card={product} /></li>{/each}
						</ul>
					{/snippet}
				</DirectoryResults>
			{:else}
				<DirectoryResults
					list={consignors}
					query={search}
					skeleton="list"
					{isVisitor}
					failedTitle={text.consignorsFailed}
					emptyMessage={text.consignorsEmpty}
					onclear={clear}
				>
					{#snippet card(consignor: DirectoryConsignorCard)}
						<ConsignorCard card={consignor} />
					{/snippet}
					{#snippet results(items: DirectoryConsignorCard[])}
						<ul class="org-row-list">
							{#each items as consignor (consignor.id)}<li>
									<ConsignorCard card={consignor} />
								</li>{/each}
						</ul>
					{/snippet}
				</DirectoryResults>
			{/if}
		{/snippet}
	</Tabs>
</FilterLayout>
