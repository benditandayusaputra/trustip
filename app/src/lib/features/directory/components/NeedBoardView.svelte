<script lang="ts">
	import { onMount, type Snippet } from 'svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { NeedBoardCard as Card } from '#lib/api/directory.js';
	import { listCategories, type Category } from '#lib/api/organizations.js';
	import { id } from '#lib/i18n/id.js';
	import { boardSearchParams, defaultBoardQuery, type BoardQuery } from '../board-query';
	import { radiusChoices } from '../query';
	import { DirectoryList } from '../state.svelte';
	import DirectoryResults from './DirectoryResults.svelte';
	import NeedBoardCard from './NeedBoardCard.svelte';

	interface Props {
		query: BoardQuery;
		canRespond: boolean;
		place: Snippet;
		onchange: (query: BoardQuery) => void;
	}

	let { query, canRespond, place, onchange }: Props = $props();

	const text = id.needBoard;
	const typingDelayMs = 300;
	const maxSearchLength = 80;
	const needs = new DirectoryList<Card>(apiClient, '/api/v1/directory/needs');
	let categories = $state<Category[]>([]);
	let typed = $derived(query.text);
	const search = $derived(boardSearchParams(query).toString());
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

<fieldset class="grid grid-cols-1 gap-4 md:grid-cols-2">
	<legend class="sr-only">{text.filters}</legend>
	<div class="md:col-span-2">
		<TextField label={text.search} type="search" maxlength={maxSearchLength} bind:value={typed} />
	</div>
	{@render place()}
	<Select
		label={id.directory.radius}
		options={radiusOptions}
		value={query.radius}
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
<DirectoryResults
	list={needs}
	query={search}
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
