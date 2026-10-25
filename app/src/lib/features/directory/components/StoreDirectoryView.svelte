<script lang="ts">
	import { onMount, type Snippet } from 'svelte';
	import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal';
	import { apiClient } from '#lib/api/client.js';
	import type { DirectoryStoreCard } from '#lib/api/directory.js';
	import { listCategories, type Category } from '#lib/api/organizations.js';
	import { directory as id } from '#lib/i18n/directory.js';
	import { defaultStoreQuery, storeSearchParams, type StoreQuery } from '../query';
	import { activeStoreFilters } from '../filter-count';
	import { DirectoryList } from '../state.svelte';
	import { forVisitor, isRandomOrder } from '../visitor';
	import DirectoryResults from './DirectoryResults.svelte';
	import StoreCard from './StoreCard.svelte';
	import StoreFilters from './StoreFilters.svelte';
	import StoreSearch from './StoreSearch.svelte';
	import StoreMap from './StoreMap.svelte';
	interface Props {
		query: StoreQuery;
		isVisitor?: boolean;
		place: Snippet;
		onchange: (query: StoreQuery) => void;
	}
	let { query, isVisitor = false, place, onchange }: Props = $props();
	const text = id.directory;
	const viewOptions = [
		{ value: 'list', label: text.listView },
		{ value: 'map', label: text.mapView }
	];
	const stores = new DirectoryList<DirectoryStoreCard>(apiClient, '/api/v1/directory/stores');
	let categories = $state<Category[]>([]);
	let filtersOpen = $state(false);
	let view = $state('list');
	const filtersId = $props.id();
	const search = $derived(storeSearchParams(forVisitor(query, isVisitor)).toString());
	const active = $derived(activeStoreFilters(query));
	function change(changes: Partial<StoreQuery>) {
		onchange({ ...query, ...changes });
	}
	onMount(() => {
		listCategories(apiClient())
			.then((loaded) => (categories = loaded))
			.catch(() => (categories = []));
	});
	$effect(() => {
		stores.shuffle = isRandomOrder(query, isVisitor);
		void stores.load(search);
	});
</script>

<div class="flex min-w-0 flex-col gap-6">
	<p class="max-w-reading text-base text-text-muted">{text.introduction}</p>
	<section class="sticky-filters" aria-label={text.searchPanel}>
		<div class="flex flex-col gap-4 rounded-card border border-border bg-surface p-4 compact:p-5">
			<div class="flex flex-wrap items-end gap-4">
				<div class="min-w-0 flex-1 basis-full md:basis-0">
					<StoreSearch {query} onchange={change} />
				</div>
				<div class="flex flex-1 items-center justify-between gap-2 md:flex-none">
					<button
						type="button"
						aria-expanded={filtersOpen}
						aria-controls={filtersId}
						onclick={() => (filtersOpen = !filtersOpen)}
						class="inline-flex min-h-12 items-center gap-2 rounded-control border border-border-strong px-4 font-semibold text-text hover:bg-surface-hover"
					>
						<SlidersHorizontal class="size-5" aria-hidden="true" />{text.filterToggle}
						{#if active > 0}<span
								class="rounded-pill bg-accent-surface px-2 py-1 text-sm whitespace-nowrap text-accent"
								>{text.activeFilters(active)}</span
							>{/if}
					</button>
					<div
						role="group"
						aria-label={text.viewLabel}
						class="flex rounded-control border border-border-strong p-1"
					>
						{#each viewOptions as { value, label } (value)}
							<button
								type="button"
								aria-pressed={view === value}
								onclick={() => (view = value)}
								class="min-h-11 rounded-control px-4 font-semibold transition-none {view === value
									? 'bg-primary text-on-accent'
									: 'text-text hover:bg-surface-hover'}">{label}</button
							>
						{/each}
					</div>
				</div>
			</div>
			<div id={filtersId} class={filtersOpen ? 'block' : 'contents'}>
				{#if filtersOpen}
					<div class="flex flex-col gap-4 border-t border-border pt-4">
						<h2 class="text-lg font-semibold text-text">{text.filters}</h2>
						<p class="text-sm text-text-muted">{text.filterHint}</p>
						<StoreFilters {query} {categories} {isVisitor} {place} onchange={change} />
						{#if active > 0}<button
								type="button"
								onclick={() => {
									onchange({ ...defaultStoreQuery(), text: query.text, sort: query.sort });
								}}
								class="min-h-11 self-start rounded-control px-4 font-semibold text-accent hover:bg-surface-hover"
								>{text.clearFilters}</button
							>{/if}
					</div>
				{/if}
			</div>
		</div>
	</section>
	{#if stores.status === 'ready'}<p role="status" class="text-sm font-semibold text-text-muted">
			{text.loadedStores(stores.items.length)}
		</p>{/if}
	<DirectoryResults
		list={stores}
		query={search}
		{isVisitor}
		failedTitle={text.failed}
		emptyMessage={text.empty}
		onclear={() => {
			onchange(defaultStoreQuery());
		}}
	>
		{#snippet card(store: DirectoryStoreCard)}<StoreCard card={store} />{/snippet}
		{#snippet results(items: DirectoryStoreCard[])}
			{#if view === 'map'}<StoreMap stores={items} />
			{:else}<ul class="list-grid">
					{#each items as store (store.id)}<li><StoreCard card={store} /></li>{/each}
				</ul>{/if}
		{/snippet}
	</DirectoryResults>
</div>
