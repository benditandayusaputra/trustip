<script lang="ts">
	import { page } from '$app/state';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import AgreementView from '#lib/features/agreements/components/AgreementView.svelte';
	import CountHistoryList from '#lib/features/counts/components/CountHistoryList.svelte';
	import DropHistoryList from '#lib/features/drops/components/DropHistoryList.svelte';
	import PartnerStock from '#lib/features/stock/components/PartnerStock.svelte';
	import PartnerPage from '#lib/features/partnerships/components/PartnerPage.svelte';
	import { id } from '#lib/i18n/id.js';

	const partnershipId = $derived(page.params.id ?? '');
	const session = getAccountSession();
	const membership = $derived(session.account?.membership);
	const recordHref = $derived(
		membership?.organization_type === 'STORE'
			? `/app/titipan/baru?mitra=${partnershipId}`
			: undefined
	);
</script>

<svelte:head>
	<title>{id.partners.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	{#key partnershipId}
		<PartnerPage
			{partnershipId}
			canAmend={membership?.role === 'OWNER'}
			isConsignor={membership?.organization_type === 'CONSIGNOR'}
		>
			{#snippet agreement()}
				<AgreementView {partnershipId} />
			{/snippet}
			{#snippet stock()}
				<PartnerStock {partnershipId} {recordHref} />
			{/snippet}
			{#snippet counts()}
				<CountHistoryList
					{partnershipId}
					countHref={recordHref ? `/app/hitung/baru?mitra=${partnershipId}` : undefined}
				/>
			{/snippet}
			{#snippet drops()}
				<DropHistoryList {partnershipId} {recordHref} />
			{/snippet}
		</PartnerPage>
	{/key}
</section>
