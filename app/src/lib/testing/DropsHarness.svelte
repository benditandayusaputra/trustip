<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import DropForm from '#lib/features/drops/components/DropForm.svelte';
	import DropHistoryList from '#lib/features/drops/components/DropHistoryList.svelte';
	import DropView from '#lib/features/drops/components/DropView.svelte';
	import PartnerChooser from '#lib/features/drops/components/PartnerChooser.svelte';

	interface Props {
		view: 'form' | 'detail' | 'choose' | 'history';
		targetId?: string;
		viewerOrganizationId?: string;
		ondone?: (documentId: string) => void;
		onqueued?: () => void;
	}

	let {
		view,
		targetId = '',
		viewerOrganizationId = '',
		ondone = () => undefined,
		onqueued = () => undefined
	}: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if view === 'form'}
	<DropForm partnershipId={targetId} onsaved={ondone} {onqueued} />
{:else if view === 'detail'}
	<DropView dropId={targetId} {viewerOrganizationId} />
{:else if view === 'history'}
	<DropHistoryList partnershipId={targetId} recordHref="/app/titipan/baru" />
{:else}
	<PartnerChooser hrefOf={(partnerId: string) => `/app/titipan/baru?mitra=${partnerId}`} />
{/if}
<Toaster />
