<script lang="ts">
	import FilterLayout from '#lib/components/ui/FilterLayout.svelte';
	import { onMount, type Snippet } from 'svelte';
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
	import ConsignorCard from './ConsignorCard.svelte';
	import ConsignorFilters from './ConsignorFilters.svelte';
	import DirectoryResults from './DirectoryResults.svelte';
	import ProductCard from './ProductCard.svelte';

	interface Props {
		query: ConsignorQuery;
		place: Snippet;
		onchange: (query: ConsignorQuery) => void;
	}

	let { query, place, onchange }: Props = $props();

	const text = id.consignorDirectory;
	const products = new DirectoryList<DirectoryProductCard>(apiClient, '/api/v1/directory/products');
	const consignors = new DirectoryList<DirectoryConsignorCard>(
		apiClient,
		'/api/v1/directory/consignors'
	);
	const tabs = Object.entries(text.tabs).map(([key, label]) => ({ key, label }));
	let categories = $state<Category[]>([]);
	const search = $derived(consignorSearchParams(query).toString());

	onMount(() => {
		listCategories(apiClient())
			.then((loaded) => (categories = loaded))
			.catch(() => (categories = []));
	});

	$effect(() => {
		const list = query.tab === 'produk' ? products : consignors;
		void list.load(search);
	});

	function change(changes: Partial<ConsignorQuery>) {
		onchange({ ...query, ...changes });
	}

	function clear() {
		onchange({ ...defaultConsignorQuery(), tab: query.tab });
	}
</script>

<FilterLayout>
	{#snippet filters()}
		<ConsignorFilters {query} {categories} {place} onchange={change} />
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
					failedTitle={text.productsFailed}
					emptyMessage={text.productsEmpty}
					onclear={clear}
				>
					{#snippet card(product: DirectoryProductCard)}
						<ProductCard card={product} />
					{/snippet}
				</DirectoryResults>
			{:else}
				<DirectoryResults
					list={consignors}
					query={search}
					failedTitle={text.consignorsFailed}
					emptyMessage={text.consignorsEmpty}
					onclear={clear}
				>
					{#snippet card(consignor: DirectoryConsignorCard)}
						<ConsignorCard card={consignor} />
					{/snippet}
				</DirectoryResults>
			{/if}
		{/snippet}
	</Tabs>
</FilterLayout>
