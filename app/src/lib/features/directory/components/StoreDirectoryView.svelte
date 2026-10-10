<script lang="ts">
	import FilterLayout from '#lib/components/ui/FilterLayout.svelte';
	import { onMount, type Snippet } from 'svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { DirectoryStoreCard } from '#lib/api/directory.js';
	import { listCategories, type Category } from '#lib/api/organizations.js';
	import { directory as id } from '#lib/i18n/directory.js';
	import { defaultStoreQuery, storeSearchParams, type StoreQuery } from '../query';
	import { DirectoryList } from '../state.svelte';
	import DirectoryResults from './DirectoryResults.svelte';
	import StoreCard from './StoreCard.svelte';
	import StoreFilters from './StoreFilters.svelte';

	interface Props {
		query: StoreQuery;
		place: Snippet;
		onchange: (query: StoreQuery) => void;
	}

	let { query, place, onchange }: Props = $props();

	const text = id.directory;
	const stores = new DirectoryList<DirectoryStoreCard>(apiClient, '/api/v1/directory/stores');
	let categories = $state<Category[]>([]);
	const search = $derived(storeSearchParams(query).toString());

	onMount(() => {
		listCategories(apiClient())
			.then((loaded) => (categories = loaded))
			.catch(() => (categories = []));
	});

	$effect(() => {
		void stores.load(search);
	});
</script>

<FilterLayout>
	{#snippet filters()}
		<StoreFilters
			{query}
			{categories}
			{place}
			onchange={(changes: Partial<StoreQuery>) => {
				onchange({ ...query, ...changes });
			}}
		/>
	{/snippet}

	<DirectoryResults
		list={stores}
		query={search}
		failedTitle={text.failed}
		emptyMessage={text.empty}
		onclear={() => {
			onchange(defaultStoreQuery());
		}}
	>
		{#snippet card(store: DirectoryStoreCard)}
			<StoreCard card={store} />
		{/snippet}
	</DirectoryResults>
</FilterLayout>
