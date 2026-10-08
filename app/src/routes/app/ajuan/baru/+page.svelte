<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Proposal } from '#lib/api/proposals.js';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import DirectProposalView from '#lib/features/proposals/components/DirectProposalView.svelte';
	import InvitationForm from '#lib/features/proposals/components/InvitationForm.svelte';
	import { id } from '#lib/i18n/id.js';

	const session = getAccountSession();
	const membership = $derived(session.account?.membership);
	const targetId = $derived(page.url.searchParams.get('sasaran') ?? '');
	const productId = $derived(page.url.searchParams.get('produk') ?? '');

	function open(proposal: Proposal) {
		void goto(`/app/ajuan/${proposal.id}`);
	}
</script>

<svelte:head>
	<title>{targetId === '' ? id.proposals.newTitle : id.directory.propose}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	{#if membership && targetId !== ''}
		<DirectProposalView
			{targetId}
			preselectedProductId={productId}
			isConsignor={membership.organization_type === 'CONSIGNOR'}
			onsent={open}
		/>
	{:else}
		<h1 class="text-xl font-semibold text-text">{id.proposals.newTitle}</h1>
		<p class="text-base text-text-muted">{id.proposals.newIntro}</p>
		{#if membership}
			<InvitationForm isConsignor={membership.organization_type === 'CONSIGNOR'} onsent={open} />
		{/if}
	{/if}
</section>
