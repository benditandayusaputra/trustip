<script lang="ts">
	import FilterLayout from '#lib/components/ui/FilterLayout.svelte';
	import { onMount, type Snippet } from 'svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { NeedBoardCard as Card } from '#lib/api/directory.js';
	import { listCategories, type Category } from '#lib/api/organizations.js';
	import { directory as id } from '#lib/i18n/directory.js';
	import { boardSearchParams, defaultBoardQuery, type BoardQuery } from '../board-query';
	import { radiusChoices } from '../query';
	import { DirectoryList } from '../state.svelte';
	import { forVisitor, isRandomOrder } from '../visitor';
	import DirectoryResults from './DirectoryResults.svelte';
	import NeedBoardCard from './NeedBoardCard.svelte';

	interface Props {
		query: BoardQuery;
		isVisitor?: boolean;
		canRespond: boolean;
		place: Snippet;
		onchange: (query: BoardQuery) => void;
	}

	let { query, isVisitor = false, canRespond, place, onchange }: Props = $props();

	const text = id.needBoard;
	const typingDelayMs = 300;
	const maxSearchLength = 80;
	const needs = new DirectoryList<Card>(apiClient, '/api/v1/directory/needs');
	let categories = $state<Category[]>([]);
	let typed = $derived(query.text);
	const search = $derived(boardSearchParams(forVisitor(query, isVisitor)).toString());
	const radiusOptions = radiusChoices.map((value) => ({ value, label: id.directory.radii[value] }));
	const sortOptions = Object.entries(text.sorts).map(([value, label]) => ({ value, label }));
	const categoryOptions = $derived([
		{ value: '', label: text.anyCategory },
		...categories.map((category) => ({ value: category.id, label: category.name }))
	]);

	onMount(() => {
		listCategories(apiClient())
			.then((loaded) => (categories = loaded))
			.catch(() => (categories = []));
	});

	$effect(() => {
		needs.shuffle = isRandomOrder(query, isVisitor);
		void needs.load(search);
	});

	$effect(() => {
		const next = typed;
		const timer = setTimeout(() => {
			if (next !== query.text) {
				change({ text: next });
			}
		}, typingDelayMs);
		return () => {
			clearTimeout(timer);
		};
	});

	function change(changes: Partial<BoardQuery>) {
		onchange({ ...query, ...changes });
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
		<fieldset class="grid grid-cols-1 items-start gap-4 md:grid-cols-2 xl:grid-cols-4">
			<legend class="sr-only">{text.filters}</legend>
			{@render place()}
			<Select
				label={id.directory.radius}
				options={radiusOptions}
				value={isVisitor ? 'all' : query.radius}
				disabled={isVisitor}
				hint={isVisitor ? id.directory.visitorRadius : undefined}
				onchange={(radius: string) => {
					change({ radius: radius as BoardQuery['radius'] });
				}}
			/>
			<Select
				label={id.directory.sort}
				options={sortOptions}
				value={query.sort}
				onchange={(sort: string) => {
					change({ sort: sort as BoardQuery['sort'] });
				}}
			/>
			<Select
				label={text.category}
				options={categoryOptions}
				value={query.categoryIds?.[0] ?? ''}
				onchange={(categoryId: string) => {
					change({ categoryIds: categoryId === '' ? null : [categoryId] });
				}}
			/>
		</fieldset>
	{/snippet}

	<DirectoryResults
		list={needs}
		query={search}
		{isVisitor}
		failedTitle={text.failed}
		emptyMessage={text.empty}
		onclear={() => {
			onchange(defaultBoardQuery());
		}}
	>
		{#snippet card(need: Card)}
			<NeedBoardCard card={need} {canRespond} />
		{/snippet}
	</DirectoryResults>
</FilterLayout>
