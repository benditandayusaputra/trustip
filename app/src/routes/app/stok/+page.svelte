<script lang="ts">
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import StockOverview from '#lib/features/stock/components/StockOverview.svelte';
	import { id } from '#lib/i18n/id.js';

	const session = getAccountSession();
	const isStore = $derived(session.account?.membership?.organization_type === 'STORE');
</script>

<svelte:head>
	<title>{id.stock.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.stock.title}</h1>
	<p class="text-base text-text-muted">{id.stock.intro}</p>
	<StockOverview
		emptyHref={isStore ? '/app/titipan/baru' : '/app/mitra'}
		emptyLabel={isStore ? id.drops.recordDrop : id.nav.partners}
	/>
</section>
