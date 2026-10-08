<script lang="ts">
	import { page } from '$app/state';
	import StockCardView from '#lib/features/stock/components/StockCardView.svelte';
	import { id } from '#lib/i18n/id.js';
</script>

<svelte:head>
	<title>{id.stock.cardTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.stock.cardTitle}</h1>
	{#key page.url.pathname}
		<StockCardView partnershipId={page.params.id ?? ''} productId={page.params.produk ?? ''} />
	{/key}
</section>
