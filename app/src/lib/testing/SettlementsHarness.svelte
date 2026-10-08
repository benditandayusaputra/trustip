<script lang="ts">
	import RunningSales from '#lib/features/settlements/components/RunningSales.svelte';
	import SettlementList from '#lib/features/settlements/components/SettlementList.svelte';
	import SettlementView from '#lib/features/settlements/components/SettlementView.svelte';

	interface Props {
		view: 'detail' | 'list' | 'running';
		targetId?: string;
	}

	let { view, targetId = '' }: Props = $props();
</script>

{#if view === 'detail'}
	<SettlementView settlementId={targetId} />
{:else if view === 'running'}
	<RunningSales partnershipId={targetId} />
{:else}
	<SettlementList partnershipId={targetId} />
{/if}
