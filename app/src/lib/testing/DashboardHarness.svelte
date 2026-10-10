<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { SalesNarrowing } from '#lib/api/reports.js';
	import ConsignorDashboard from '#lib/features/dashboard/components/ConsignorDashboard.svelte';
	import StoreDashboard from '#lib/features/dashboard/components/StoreDashboard.svelte';

	interface Props {
		isOwner?: boolean;
		side?: 'CONSIGNOR' | 'STORE';
	}

	let { isOwner = true, side = 'CONSIGNOR' }: Props = $props();

	const range = { from: '2026-10-22', to: '2026-11-20' };
</script>

{#if side === 'STORE'}
	<StoreDashboard {isOwner} {range} />
{:else}
	<ConsignorDashboard {isOwner} {range}>
		{#snippet chart(narrowing: SalesNarrowing, controls: Snippet)}
			{@render controls()}
			<p>chart {narrowing.partnerId ?? narrowing.productId ?? 'all'}</p>
		{/snippet}
	</ConsignorDashboard>
{/if}
