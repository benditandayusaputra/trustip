<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Handshake from '@lucide/svelte/icons/handshake';
	import { page } from '$app/state';
	import StockCardView from '#lib/features/stock/components/StockCardView.svelte';
	import { stock as id } from '#lib/i18n/stock.js';
</script>

<svelte:head>
	<title>{id.stock.cardTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<PageHeading icon={Handshake}>{id.stock.cardTitle}</PageHeading>
	{#key page.url.pathname}
		<StockCardView partnershipId={page.params.id ?? ''} productId={page.params.produk ?? ''} />
	{/key}
</section>
