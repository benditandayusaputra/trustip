<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import PublicConsignorProfile from '#lib/features/directory/components/PublicConsignorProfile.svelte';
	import PublicNeedPage from '#lib/features/directory/components/PublicNeedPage.svelte';
	import PublicProductPage from '#lib/features/directory/components/PublicProductPage.svelte';
	import PublicStoreProfile from '#lib/features/directory/components/PublicStoreProfile.svelte';
	import type { PublicConsignor, PublicNeed, PublicProduct, PublicStore } from '#lib/api/public.js';

	interface Props {
		store?: PublicStore;
		consignor?: PublicConsignor;
		product?: PublicProduct;
		need?: PublicNeed;
		locate?: () => Promise<{ latitude: number; longitude: number }>;
	}

	let {
		store,
		consignor,
		product,
		need,
		locate = () => Promise.reject(new Error('denied'))
	}: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if store}
	<PublicStoreProfile {store} shareUrl="https://trustip.test/t/x" />
{/if}
{#if consignor}
	<PublicConsignorProfile {consignor} shareUrl="https://trustip.test/u/x" />
{/if}
{#if product}
	<PublicProductPage
		{product}
		consignorSlug="roti-mak-ijah"
		{locate}
		shareUrl="https://trustip.test/p/x"
	/>
{/if}
{#if need}
	<PublicNeedPage {need} shareUrl="https://trustip.test/kebutuhan/x" />
{/if}
<Toaster />
