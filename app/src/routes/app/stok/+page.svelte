<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Warehouse from '@lucide/svelte/icons/warehouse';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import StockOverview from '#lib/features/stock/components/StockOverview.svelte';
	import { common as commonText } from '#lib/i18n/common.js';
	import { drops as dropsText } from '#lib/i18n/drops.js';
	import { stock as stockText } from '#lib/i18n/stock.js';

	const id = { ...commonText, ...dropsText, ...stockText };

	const session = getAccountSession();
	const isStore = $derived(session.account?.membership?.organization_type === 'STORE');
</script>

<svelte:head>
	<title>{id.stock.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<PageHeading icon={Warehouse}>{id.stock.title}</PageHeading>
	<p class="text-base text-text-muted">{id.stock.intro}</p>
	<StockOverview
		emptyHref={isStore ? '/app/titipan/baru' : '/app/mitra'}
		emptyLabel={isStore ? id.drops.recordDrop : id.nav.partners}
	/>
</section>
