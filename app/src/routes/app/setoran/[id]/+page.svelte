<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Wallet from '@lucide/svelte/icons/wallet';
	import { page } from '$app/state';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import SettlementView from '#lib/features/settlements/components/SettlementView.svelte';
	import { settlements as id } from '#lib/i18n/settlements.js';

	const session = getAccountSession();
	const viewer = $derived({
		organizationId: session.account?.membership?.organization_id ?? '',
		isOwner: session.account?.membership?.role === 'OWNER'
	});
</script>

<svelte:head>
	<title>{id.settlements.detailTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<PageHeading backHref="/app/laporan" icon={Wallet}>{id.settlements.detailTitle}</PageHeading>
	{#key page.params.id}
		<SettlementView settlementId={page.params.id ?? ''} {viewer} />
	{/key}
</section>
