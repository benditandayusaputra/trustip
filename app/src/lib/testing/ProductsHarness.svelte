<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import type { Product } from '#lib/api/products.js';
	import ProductCatalog from '#lib/features/products/components/ProductCatalog.svelte';
	import ProductDetail from '#lib/features/products/components/ProductDetail.svelte';
	import ProductForm from '#lib/features/products/components/ProductForm.svelte';

	interface Props {
		view: 'catalog' | 'form' | 'detail';
		productId?: string;
		onsaved?: (product: Product) => void;
	}

	let { view, productId = '', onsaved = () => undefined }: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if view === 'catalog'}
	<ProductCatalog />
{:else if view === 'form'}
	<ProductForm product={null} {onsaved} />
{:else}
	<ProductDetail {productId} />
{/if}
<Toaster />
