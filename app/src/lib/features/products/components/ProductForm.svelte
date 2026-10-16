<script lang="ts">
	import { tick, untrack } from 'svelte';
	import ProductFields from './ProductFields.svelte';
	import Button from '#lib/components/ui/Button.svelte';
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

<form
	bind:this={formElement}
	onsubmit={handleSubmit}
	novalidate
	class="surface-card flex flex-col gap-5 p-5 compact:p-6"
>
	<fieldset {disabled} class="flex flex-col gap-5">
		<ProductFields {editor} {categoryOptions} {categoriesLoading} />
		<div class="form-footer">
			<a
				href="/app/produk"
				class="inline-flex min-h-12 items-center justify-center rounded-control border border-border px-5 text-sm font-semibold text-text"
				>{text.backToCatalog}</a
			>
			<Button type="submit" busy={editor.busy}>{text.save}</Button>
		</div>
	</fieldset>
</form>
