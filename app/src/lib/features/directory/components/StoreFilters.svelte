<script lang="ts">
	import type { Snippet } from 'svelte';
	import Checkbox from '#lib/components/ui/Checkbox.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import type { Category } from '#lib/api/organizations.js';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { organization as organizationText } from '#lib/i18n/organization.js';
	import { anyCategory, radiusChoices, sortChoices, type StoreQuery } from '../query';

	const id = { ...directoryText, ...organizationText };

	interface Props {
		query: StoreQuery;
		categories: Category[];
		place: Snippet;
		onchange: (changes: Partial<StoreQuery>) => void;
	}

	let { query, categories, place, onchange }: Props = $props();

	const text = id.directory;
	const typingDelayMs = 300;
	const maxSearchLength = 80;
	const ownCategories = 'own';
	let typed = $derived(query.text);
	const radiusOptions = radiusChoices.map((value) => ({ value, label: text.radii[value] }));
	const sortOptions = sortChoices.map((value) => ({ value, label: text.sorts[value] }));
	const storeTypeOptions = [
		{ value: '', label: text.anyStoreType },
		...Object.entries(id.storeProfile.storeTypes).map(([value, label]) => ({ value, label }))
	];
	const modelOptions = [
		{ value: '', label: text.anyModel },
		...Object.entries(text.models).map(([value, label]) => ({ value, label }))
	];
	const categoryOptions = $derived([
		{ value: ownCategories, label: text.ownCategories },
		{ value: anyCategory, label: text.anyCategory },
		...categories.map((category) => ({ value: category.id, label: category.name }))
	]);
	const categoryValue = $derived(query.categoryIds?.[0] ?? ownCategories);

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
</script>

<fieldset class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-1">
	<legend class="sr-only">{text.filters}</legend>
	<div class="md:col-span-2 xl:col-span-1">
		<TextField label={text.search} type="search" maxlength={maxSearchLength} bind:value={typed} />
	</div>
	{@render place()}
	<Select
		label={text.radius}
		options={radiusOptions}
		value={query.radius}
		onchange={(radius: string) => {
			onchange({ radius: radius as StoreQuery['radius'] });
		}}
	/>
	<Select
		label={text.sort}
		options={sortOptions}
		value={query.sort}
		onchange={(sort: string) => {
			onchange({ sort: sort as StoreQuery['sort'] });
		}}
	/>
	<Select
		label={text.storeType}
		options={storeTypeOptions}
		value={query.storeType}
		onchange={(storeType: string) => {
			onchange({ storeType });
		}}
	/>
	<Select
		label={text.category}
		options={categoryOptions}
		value={categoryValue}
		onchange={(choice: string) => {
			onchange({ categoryIds: choice === ownCategories ? null : [choice] });
		}}
	/>
	<Select
		label={text.pricingModel}
		options={modelOptions}
		value={query.pricingModel}
		onchange={(pricingModel: string) => {
			onchange({ pricingModel });
		}}
	/>
	<div class="flex flex-col gap-2">
		<Checkbox
			checked={query.withoutPermit}
			onchange={(withoutPermit: boolean) => {
				onchange({ withoutPermit });
			}}>{text.withoutPermit}</Checkbox
		>
		<Checkbox
			checked={query.withoutHalal}
			onchange={(withoutHalal: boolean) => {
				onchange({ withoutHalal });
			}}>{text.withoutHalal}</Checkbox
		>
		<Checkbox
			checked={query.hasOpenNeed}
			onchange={(hasOpenNeed: boolean) => {
				onchange({ hasOpenNeed });
			}}>{text.hasOpenNeed}</Checkbox
		>
	</div>
</fieldset>
