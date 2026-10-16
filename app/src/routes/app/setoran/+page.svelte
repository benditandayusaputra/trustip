<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Wallet from '@lucide/svelte/icons/wallet';
	import SettlementList from '#lib/features/settlements/components/SettlementList.svelte';
	import { settlements as id } from '#lib/i18n/settlements.js';
</script>

<svelte:head>
	<title>{id.settlements.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<PageHeading icon={Wallet}>{id.settlements.title}</PageHeading>
	<SettlementList />
</section>
