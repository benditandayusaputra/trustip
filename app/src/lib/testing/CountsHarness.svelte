<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import CountForm from '#lib/features/counts/components/CountForm.svelte';
	import CountHistoryList from '#lib/features/counts/components/CountHistoryList.svelte';
	import CountView from '#lib/features/counts/components/CountView.svelte';

	interface Props {
		view: 'form' | 'recap' | 'history';
		targetId?: string;
		viewerOrganizationId?: string;
		ondone?: (documentId: string) => void;
		onqueued?: () => void;
		revisionOf?: string;
	}

	let {
		view,
		targetId = '',
		viewerOrganizationId = '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d',
		ondone = () => undefined,
		onqueued = () => undefined,
		revisionOf = ''
	}: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if view === 'form'}
	<CountForm
		partnershipId={targetId}
		dropHref="/app/titipan/baru"
		onsaved={ondone}
		{onqueued}
		{revisionOf}
	/>
{:else if view === 'history'}
	<CountHistoryList partnershipId={targetId} countHref="/app/hitung/baru" />
{:else}
	<CountView countId={targetId} {viewerOrganizationId} />
{/if}
<Toaster />
