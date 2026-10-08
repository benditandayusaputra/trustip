<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import PartnerList from '#lib/features/partnerships/components/PartnerList.svelte';
	import PartnerPage from '#lib/features/partnerships/components/PartnerPage.svelte';

	interface Props {
		view: 'list' | 'page';
		partnershipId?: string;
		isConsignor?: boolean;
	}

	let { view, partnershipId = '', isConsignor = true }: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if view === 'list'}
	<PartnerList />
{:else}
	<PartnerPage {partnershipId} canAmend={true} {isConsignor}>
		{#snippet agreement()}
			<p>Kesepakatan tampil di sini</p>
		{/snippet}
	</PartnerPage>
{/if}
<Toaster />
