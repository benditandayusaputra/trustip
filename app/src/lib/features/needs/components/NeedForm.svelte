<script lang="ts">
	import { onMount, tick } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import FieldSkeleton from '#lib/components/ui/FieldSkeleton.svelte';
	import MoneyField from '#lib/components/ui/MoneyField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import TextArea from '#lib/components/ui/TextArea.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Need } from '#lib/api/needs.js';
	import { listCategories, type Category } from '#lib/api/organizations.js';
	import { needs as id } from '#lib/i18n/needs.js';
	import { NeedComposer } from '../state.svelte';

	interface Props {
		onposted: (need: Need) => void;
	}

	let { onposted }: Props = $props();

	const text = id.needs;
	const composer = new NeedComposer(apiClient);
	const toaster = getToaster();
	const maxTitleLength = 80;
	let categories = $state<Category[]>([]);
	let categoriesLoading = $state(true);
	let quantityText = $state('');
	let formElement: HTMLFormElement | undefined = $state();
	const categoryOptions = $derived([
		{ value: '', label: text.chooseCategory },
		...categories.map((category) => ({ value: category.id, label: category.name }))
	]);
	const priceError = $derived(
		composer.errors.maxRetailPrice ? { error: composer.errors.maxRetailPrice } : {}
	);
	const modelOptions = [
		{ value: '', label: text.anyModel },
		...Object.entries(text.models).map(([value, label]) => ({ value, label }))
	];

	onMount(() => {
		listCategories(apiClient())
			.then((loaded) => (categories = loaded))
			.catch(() => (categories = []))
			.finally(() => (categoriesLoading = false));
	});

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const quantity = Number(quantityText);
		composer.values.weeklyQuantity = quantityText === '' ? undefined : quantity;
		const outcome = await composer.post();
		if (outcome.kind === 'posted') {
			toaster.show({ tone: 'success', message: text.posted });
			onposted(outcome.need);
		} else if (outcome.kind === 'failed') {
			toaster.show({ tone: 'danger', message: outcome.message });
		} else {
			await tick();
			formElement?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
		}
	}
</script>

<form
	bind:this={formElement}
	onsubmit={handleSubmit}
	novalidate
	class="surface-card flex flex-col gap-6 p-5 compact:p-6"
>
	<TextField
		label={text.title}
		hint={text.titleHint}
		maxlength={maxTitleLength}
		bind:value={composer.values.title}
		error={composer.errors.title}
	/>
	{#if categoriesLoading}
		<FieldSkeleton label={text.category} />
	{:else}
		<Select
			label={text.category}
			options={categoryOptions}
			bind:value={composer.values.categoryId}
			error={composer.errors.categoryId}
		/>
	{/if}
	<TextArea
		label={text.description}
		hint={text.descriptionHint}
		bind:value={composer.values.description}
		error={composer.errors.description}
	/>
	<TextField
		label={text.weeklyQuantity}
		inputmode="numeric"
		hint={text.optional}
		bind:value={quantityText}
		error={composer.errors.weeklyQuantity}
	/>
	<MoneyField
		label={text.maxRetailPrice}
		hint={text.optional}
		bind:value={composer.values.maxRetailPrice}
		{...priceError}
	/>
	<Select
		label={text.model}
		options={modelOptions}
		value={composer.values.model}
		onchange={(model: string) => {
			composer.values.model = model as typeof composer.values.model;
		}}
	/>
	{#if composer.values.model !== 'FIXED_DEPOSIT'}
		<TextField
			label={text.commission}
			inputmode="decimal"
			hint={text.optional}
			bind:value={composer.values.commissionPercent}
			error={composer.errors.commissionPercent}
		/>
	{/if}
	<Button type="submit" busy={composer.busy}>{text.post}</Button>
</form>
