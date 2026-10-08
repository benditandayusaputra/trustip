<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import type { Need } from '#lib/api/needs.js';
	import NeedDetail from '#lib/features/needs/components/NeedDetail.svelte';
	import NeedForm from '#lib/features/needs/components/NeedForm.svelte';
	import NeedList from '#lib/features/needs/components/NeedList.svelte';

	interface Props {
		view: 'form' | 'list' | 'detail';
		needId?: string;
		onposted?: (need: Need) => void;
	}

	let { view, needId = '', onposted = () => undefined }: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if view === 'form'}
	<NeedForm {onposted} />
{:else if view === 'list'}
	<NeedList />
{:else}
	<NeedDetail {needId} />
{/if}
<Toaster />
