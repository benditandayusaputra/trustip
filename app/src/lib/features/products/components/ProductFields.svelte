<script lang="ts">
	import FormSection from '#lib/components/ui/FormSection.svelte';
	import FieldSkeleton from '#lib/components/ui/FieldSkeleton.svelte';
	import ImageField from '#lib/components/ui/ImageField.svelte';
	import MoneyField from '#lib/components/ui/MoneyField.svelte';
	import NumberField from '#lib/components/ui/NumberField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import Switch from '#lib/components/ui/Switch.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import TextArea from '#lib/components/ui/TextArea.svelte';
	import { organization as organizationText } from '#lib/i18n/organization.js';
	import { products as productsText } from '#lib/i18n/products.js';
	import type { ProductEditor } from '../editor.svelte';
	interface Props {
		editor: ProductEditor;
		categoryOptions: { value: string; label: string }[];
		categoriesLoading: boolean;
	}
	let { editor, categoryOptions, categoriesLoading }: Props = $props();
	const id = { ...organizationText, ...productsText };
	const text = id.products;
	const maxShelfLifeDays = 730;
</script>

<FormSection title={text.identityTitle} description={text.identityHint}>
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
	<div class="grid gap-5 md:grid-cols-2">
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
	</div>
</FormSection>
<FormSection title={text.pricingTitle} description={text.pricingHint}>
	<div class="grid gap-5 md:grid-cols-2">
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
	</div>
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
</FormSection>
<FormSection title={text.detailsTitle} description={text.detailsHint}>
	<TextField
		label={text.netContent}
		name="net_content"
		hint={text.netContentHint}
		bind:value={editor.values.netContent}
		error={editor.errors.netContent}
	/>
	<TextArea
		label={text.description}
		name="description"
		bind:value={editor.values.description}
		error={editor.errors.description}
	/>
	<TextArea
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
</FormSection>
