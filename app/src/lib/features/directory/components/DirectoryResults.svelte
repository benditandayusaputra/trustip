<script lang="ts" generics="T extends { id: string }">
	import { onMount, type Snippet } from 'svelte';
	import SearchX from '@lucide/svelte/icons/search-x';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { directory as id } from '#lib/i18n/directory.js';
	import { watchOnline } from '#lib/offline/network.js';
	import type { DirectoryList } from '../state.svelte';

	interface Props {
		list: DirectoryList<T>;
		query: string;
		failedTitle: string;
		emptyMessage: string;
		card: Snippet<[T]>;
		results?: Snippet<[T[]]>;
		onclear: () => void;
	}

	let { list, query, failedTitle, emptyMessage, card, results, onclear }: Props = $props();

	const text = id.directory;
	let isOnline = $state(true);
	const search = $derived(list.search);

	onMount(() => watchOnline((online) => (isOnline = online)));
</script>

<div class="flex min-w-0 flex-col gap-4">
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
	{#if list.status === 'loading'}
		<ContentSkeleton variant="cards" />
	{:else if list.status === 'failed'}
		<Alert title={failedTitle}>
			{list.failure}
			{#snippet action()}
				<Button variant="secondary" onclick={() => void list.load(query)}>{text.retry}</Button>
			{/snippet}
		</Alert>
	{:else if list.items.length === 0}
		<EmptyState icon={SearchX} message={emptyMessage}>
			{#snippet action()}
				<Button variant="secondary" onclick={onclear}>{text.clearFilters}</Button>
			{/snippet}
		</EmptyState>
	{:else}
		{#if results}
			{@render results(list.items)}
		{:else}
			<ul class="grid grid-cols-1 gap-4 lg:grid-cols-2">
				{#each list.items as item (item.id)}
					<li>{@render card(item)}</li>
				{/each}
			</ul>
		{/if}
		{#if list.nextCursor}
			<Button variant="secondary" busy={list.busy} onclick={() => void list.loadMore()}>
				{text.loadMore}
			</Button>
		{/if}
	{/if}
</div>
