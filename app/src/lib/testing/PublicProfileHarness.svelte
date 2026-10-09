<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import PublicConsignorProfile from '#lib/features/directory/components/PublicConsignorProfile.svelte';
	import PublicProductPage from '#lib/features/directory/components/PublicProductPage.svelte';
	import PublicStoreProfile from '#lib/features/directory/components/PublicStoreProfile.svelte';
	import type { PublicConsignor, PublicProduct, PublicStore } from '#lib/api/public.js';

	interface Props {
		store?: PublicStore;
		consignor?: PublicConsignor;
		product?: PublicProduct;
		locate?: () => Promise<{ latitude: number; longitude: number }>;
	}

	let {
		store,
		consignor,
		product,
		locate = () => Promise.reject(new Error('denied'))
	}: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if store}
	<PublicStoreProfile {store} />
{/if}
{#if consignor}
	<PublicConsignorProfile {consignor} />
{/if}
{#if product}
	<PublicProductPage {product} consignorSlug="roti-mak-ijah" {locate} />
{/if}
<Toaster />
