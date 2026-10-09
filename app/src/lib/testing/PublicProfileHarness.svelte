<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import PublicConsignorProfile from '#lib/features/directory/components/PublicConsignorProfile.svelte';
	import PublicStoreProfile from '#lib/features/directory/components/PublicStoreProfile.svelte';
	import type { PublicConsignor, PublicStore } from '#lib/api/public.js';

	interface Props {
		store?: PublicStore;
		consignor?: PublicConsignor;
	}

	let { store, consignor }: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if store}
	<PublicStoreProfile {store} />
{/if}
{#if consignor}
	<PublicConsignorProfile {consignor} />
{/if}
<Toaster />
