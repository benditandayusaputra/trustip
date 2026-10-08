<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import type { Count } from '#lib/api/counts.js';
	import CountForm from '#lib/features/counts/components/CountForm.svelte';
	import CountView from '#lib/features/counts/components/CountView.svelte';

	interface Props {
		view: 'form' | 'recap';
		targetId?: string;
		ondone?: (count: Count) => void;
	}

	let { view, targetId = '', ondone = () => undefined }: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if view === 'form'}
	<CountForm partnershipId={targetId} dropHref="/app/titipan/baru" onsaved={ondone} />
{:else}
	<CountView countId={targetId} isStore={true} />
{/if}
<Toaster />
