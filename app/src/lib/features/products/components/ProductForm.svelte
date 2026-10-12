<script lang="ts">
	import { tick, untrack } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import FieldSkeleton from '#lib/components/ui/FieldSkeleton.svelte';
	import ImageField from '#lib/components/ui/ImageField.svelte';
	import MoneyField from '#lib/components/ui/MoneyField.svelte';
	import NumberField from '#lib/components/ui/NumberField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import Switch from '#lib/components/ui/Switch.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Product } from '#lib/api/products.js';
	import { organization as organizationText } from '#lib/i18n/organization.js';
	import { products as productsText } from '#lib/i18n/products.js';
	import { ProductEditor } from '../editor.svelte';

	const id = { ...organizationText, ...productsText };

	interface Props {
		product: Product | null;
		disabled?: boolean;
		onsaved: (product: Product) => void;
	}

	let { product, disabled = false, onsaved }: Props = $props();

	const text = id.products;
	const editor = new ProductEditor(
		apiClient,
		untrack(() => product)
	);
	const toaster = getToaster();
	const maxShelfLifeDays = 730;
	const categoryOptions = $derived([
		{ value: '', label: id.onboarding.choose },
		...editor.categories.map((category) => ({ value: category.id, label: category.name }))
	]);
	let formElement: HTMLFormElement | undefined = $state();
	let categoriesLoading = $state(true);

	$effect(() => {
		void editor.loadCategories().finally(() => (categoriesLoading = false));
	});

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const outcome = await editor.save();
		if (outcome.kind === 'saved') {
			toaster.show({ tone: 'success', message: product ? text.saved : text.created });
			onsaved(outcome.product);
		} else if (outcome.kind === 'failed') {
			toaster.show({ tone: 'danger', message: outcome.message });
		} else {
			await tick();
			formElement?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
		}
	}
</script>

<form bind:this={formElement} onsubmit={handleSubmit} novalidate class="flex flex-col gap-5">
	<fieldset {disabled} class="flex flex-col gap-5">
		<TextField
			label={text.name}
			name="name"
			bind:value={editor.values.name}
			error={editor.errors.name}
		/>
		{#if categoriesLoading}
			<FieldSkeleton label={text.category} />
		{:else}
			<Select
				label={text.category}
				name="category_id"
				options={categoryOptions}
				bind:value={editor.values.categoryId}
				error={editor.errors.categoryId}
			/>
		{/if}
		<TextField
			label={text.sku}
			name="sku"
			hint={text.skuHint}
			bind:value={editor.values.sku}
			error={editor.errors.sku}
		/>
		<TextField
			label={text.unit}
			name="unit"
			hint={text.unitHint}
			bind:value={editor.values.unit}
			error={editor.errors.unit}
		/>
		<MoneyField
			label={text.suggestedRetail}
			name="suggested_retail_price"
			hint={text.suggestedHint}
			bind:value={editor.values.retailPrice}
			{...editor.errors.retailPrice && { error: editor.errors.retailPrice }}
		/>
		<MoneyField
			label={text.suggestedDeposit}
			name="suggested_deposit_price"
			bind:value={editor.values.depositPrice}
			{...editor.errors.depositPrice && { error: editor.errors.depositPrice }}
		/>
		<Switch label={text.hasShelfLife} bind:checked={editor.values.hasShelfLife} />
		{#if editor.values.hasShelfLife}
			<NumberField
				label={text.shelfLife}
				name="shelf_life_days"
				hint={text.shelfLifeHint}
				min={1}
				max={maxShelfLifeDays}
				bind:value={editor.values.shelfLifeDays}
				{...editor.errors.shelfLifeDays && { error: editor.errors.shelfLifeDays }}
			/>
		{/if}
		<TextField
			label={text.netContent}
			name="net_content"
			hint={text.netContentHint}
			bind:value={editor.values.netContent}
			error={editor.errors.netContent}
		/>
		<TextField
			label={text.description}
			name="description"
			bind:value={editor.values.description}
			error={editor.errors.description}
		/>
		<TextField
			label={text.composition}
			name="composition"
			bind:value={editor.values.composition}
			error={editor.errors.composition}
		/>
		<ImageField
			label={text.photo}
			chooseLabel={text.choosePhoto}
			imageUrl={editor.photo?.url}
			imageAlt={text.photoAlt(editor.values.name)}
			hint={id.settings.logoHint}
			error={editor.errors.photo}
			busy={editor.isUploading}
			onselect={(file: File) => void editor.choosePhoto(file)}
		/>
		<Button type="submit" busy={editor.busy}>{text.save}</Button>
	</fieldset>
</form>
