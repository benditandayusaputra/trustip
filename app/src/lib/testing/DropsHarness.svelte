<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import type { Drop } from '#lib/api/drops.js';
	import DropForm from '#lib/features/drops/components/DropForm.svelte';
	import DropView from '#lib/features/drops/components/DropView.svelte';
	import PartnerChooser from '#lib/features/drops/components/PartnerChooser.svelte';

	interface Props {
		view: 'form' | 'detail' | 'choose';
		targetId?: string;
		viewerOrganizationId?: string;
		ondone?: (drop: Drop) => void;
	}

	let {
		view,
		targetId = '',
		viewerOrganizationId = '',
		ondone = () => undefined
	}: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if view === 'form'}
	<DropForm partnershipId={targetId} onsaved={ondone} />
{:else if view === 'detail'}
	<DropView dropId={targetId} {viewerOrganizationId} />
{:else}
	<PartnerChooser hrefOf={(partnerId: string) => `/app/titipan/baru?mitra=${partnerId}`} />
{/if}
<Toaster />
