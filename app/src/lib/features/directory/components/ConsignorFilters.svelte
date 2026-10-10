<script lang="ts">
	import type { Snippet } from 'svelte';
	import Checkbox from '#lib/components/ui/Checkbox.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import type { Category } from '#lib/api/organizations.js';
	import { directory as id } from '#lib/i18n/directory.js';
	import type { ConsignorQuery } from '../consignor-query';
	import { priceRangeOf, priceRanges } from '../price-ranges';
	import { radiusChoices, sortChoices } from '../query';

	interface Props {
		query: ConsignorQuery;
		categories: Category[];
		place: Snippet;
		onchange: (changes: Partial<ConsignorQuery>) => void;
	}

	let { query, categories, place, onchange }: Props = $props();

	const text = id.consignorDirectory;
	const shared = id.directory;
	const typingDelayMs = 300;
	const maxSearchLength = 80;
	let typed = $derived(query.text);
	const radiusOptions = radiusChoices.map((value) => ({ value, label: shared.radii[value] }));
	const sortOptions = sortChoices.map((value) => ({ value, label: shared.sorts[value] }));
	const priceOptions = priceRanges.map((range) => ({
		value: range.key,
		label: text.priceRanges[range.key as keyof typeof text.priceRanges]
	}));
	const shelfLifeOptions = Object.entries(text.shelfLives).map(([value, label]) => ({
		value: value === 'any' ? '' : value,
		label
	}));
	const categoryOptions = $derived([
		{ value: '', label: text.anyCategory },
		...categories.map((category) => ({ value: category.id, label: category.name }))
	]);
	const checks = $derived([
		{ key: 'hasPermit', label: text.hasPermit, checked: query.hasPermit },
		{ key: 'hasHalal', label: text.hasHalal, checked: query.hasHalal },
		{ key: 'canSample', label: text.canSample, checked: query.canSample },
		{ key: 'deliversToMe', label: text.deliversToMe, checked: query.deliversToMe }
	] as const);

	$effect(() => {
		const next = typed;
		const timer = setTimeout(() => {
			if (next !== query.text) {
				onchange({ text: next });
			}
		}, typingDelayMs);
		return () => {
			clearTimeout(timer);
		};
	});

	function choosePriceRange(key: string) {
		const range = priceRanges.find((candidate) => candidate.key === key);
		onchange({ minPrice: range?.minPrice ?? '', maxPrice: range?.maxPrice ?? '' });
	}
</script>

<fieldset class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-1">
	<legend class="sr-only">{text.filters}</legend>
	<div class="md:col-span-2 xl:col-span-1">
		<TextField label={text.search} type="search" maxlength={maxSearchLength} bind:value={typed} />
	</div>
	{@render place()}
	<Select
		label={shared.radius}
		options={radiusOptions}
		value={query.radius}
		onchange={(radius: string) => {
			onchange({ radius: radius as ConsignorQuery['radius'] });
		}}
	/>
	<Select
		label={shared.sort}
		options={sortOptions}
		value={query.sort}
		onchange={(sort: string) => {
			onchange({ sort: sort as ConsignorQuery['sort'] });
		}}
	/>
	<Select
		label={text.category}
		options={categoryOptions}
		value={query.categoryIds?.[0] ?? ''}
		onchange={(categoryId: string) => {
			onchange({ categoryIds: categoryId === '' ? null : [categoryId] });
		}}
	/>
	<Select
		label={text.priceRange}
		options={priceOptions}
		value={priceRangeOf(query.minPrice, query.maxPrice)}
		onchange={choosePriceRange}
	/>
	<Select
		label={text.shelfLife}
		options={shelfLifeOptions}
		value={query.minShelfLifeDays}
		onchange={(minShelfLifeDays: string) => {
			onchange({ minShelfLifeDays });
		}}
	/>
	<div class="flex flex-col gap-2">
		{#each checks as check (check.key)}
			<Checkbox
				checked={check.checked}
				onchange={(checked: boolean) => {
					onchange({ [check.key]: checked });
				}}>{check.label}</Checkbox
			>
		{/each}
	</div>
</fieldset>
