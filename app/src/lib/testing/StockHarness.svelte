<script lang="ts">
	import PartnerStock from '#lib/features/stock/components/PartnerStock.svelte';
	import StockCardView from '#lib/features/stock/components/StockCardView.svelte';
	import StockOverview from '#lib/features/stock/components/StockOverview.svelte';

	interface Props {
		view: 'overview' | 'partner' | 'card';
		partnershipId?: string;
		productId?: string;
	}

	let { view, partnershipId = '', productId = '' }: Props = $props();
</script>

{#if view === 'overview'}
	<StockOverview emptyHref="/app/titipan/baru" emptyLabel="Catat titipan" />
{:else if view === 'partner'}
	<PartnerStock {partnershipId} recordHref="/app/titipan/baru" />
{:else}
	<StockCardView {partnershipId} {productId} />
{/if}
