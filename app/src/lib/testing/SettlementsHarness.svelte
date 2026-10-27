<script lang="ts">
	import type { SettlementNarrowing } from '#lib/api/settlements.js';
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import RunningSales from '#lib/features/settlements/components/RunningSales.svelte';
	import SettlementList from '#lib/features/settlements/components/SettlementList.svelte';
	import SettlementView from '#lib/features/settlements/components/SettlementView.svelte';

	interface Props {
		view: 'detail' | 'list' | 'running';
		targetId?: string;
		viewerOrganizationId?: string;
		narrowing?: SettlementNarrowing;
	}

	let {
		view,
		targetId = '',
		viewerOrganizationId = '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d',
		narrowing
	}: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if view === 'detail'}
	<SettlementView
		settlementId={targetId}
		viewer={{ organizationId: viewerOrganizationId, isOwner: true }}
	/>
{:else if view === 'running'}
	<RunningSales partnershipId={targetId} />
{:else}
	{#if narrowing}
		<SettlementList partnershipId={targetId} {narrowing} heading="Daftar Tagihan Setoran" />
	{:else}
		<SettlementList partnershipId={targetId} />
	{/if}
{/if}
<Toaster />
