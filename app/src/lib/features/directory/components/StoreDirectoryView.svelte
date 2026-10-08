<script lang="ts">
	import { onMount, type Snippet } from 'svelte';
	import SearchX from '@lucide/svelte/icons/search-x';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { listCategories, type Category } from '#lib/api/organizations.js';
	import { id } from '#lib/i18n/id.js';
	import { watchOnline } from '#lib/offline/network.js';
	import { defaultStoreQuery, storeSearchParams, type StoreQuery } from '../query';
	import { StoreDirectory } from '../state.svelte';
	import StoreCard from './StoreCard.svelte';
	import StoreFilters from './StoreFilters.svelte';

	interface Props {
		query: StoreQuery;
		place: Snippet;
		onchange: (query: StoreQuery) => void;
	}

	let { query, place, onchange }: Props = $props();

	const text = id.directory;
	const directory = new StoreDirectory(apiClient);
	let categories = $state<Category[]>([]);
	let isOnline = $state(true);
	const search = $derived(directory.search);

	onMount(() => watchOnline((online) => (isOnline = online)));
	onMount(() => {
		listCategories(apiClient())
			.then((loaded) => (categories = loaded))
			.catch(() => (categories = []));
	});

	$effect(() => {
		void directory.load(storeSearchParams(query).toString());
	});

	function change(changes: Partial<StoreQuery>) {
		onchange({ ...query, ...changes });
	}
</script>

<StoreFilters {query} {categories} {place} onchange={change} />
{#if !isOnline}
	<p role="status" class="text-sm text-warning">{text.offline}</p>
{/if}
{#if search?.is_expanded}
	<p role="status" class="text-sm text-accent">{text.expanded(search.effective_radius_km)}</p>
{/if}
{#if search && !search.origin.is_known}
	<p role="status" class="text-sm text-text-muted">
		{text.noOrigin}
		<a href="/app/pengaturan" class="font-semibold text-accent">{text.setLocation}</a>
	</p>
{/if}
{#if directory.status === 'loading'}
	<Skeleton />
{:else if directory.status === 'failed'}
	<Alert title={text.failed}>
		{directory.failure}
		{#snippet action()}
			<Button
				variant="secondary"
				onclick={() => void directory.load(storeSearchParams(query).toString())}
			>
				{text.retry}
			</Button>
		{/snippet}
	</Alert>
{:else if directory.items.length === 0}
	<EmptyState icon={SearchX} message={text.empty}>
		{#snippet action()}
			<Button
				variant="secondary"
				onclick={() => {
					onchange(defaultStoreQuery());
				}}
			>
				{text.clearFilters}
			</Button>
		{/snippet}
	</EmptyState>
{:else}
	<ul class="grid grid-cols-1 gap-4 lg:grid-cols-2">
		{#each directory.items as card (card.id)}
			<li><StoreCard {card} /></li>
		{/each}
	</ul>
	{#if directory.nextCursor}
		<Button variant="secondary" busy={directory.busy} onclick={() => void directory.loadMore()}>
			{text.loadMore}
		</Button>
	{/if}
{/if}
