<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import FileSignature from '@lucide/svelte/icons/file-signature';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Proposal } from '#lib/api/proposals.js';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import DirectProposalView from '#lib/features/proposals/components/DirectProposalView.svelte';
	import InvitationForm from '#lib/features/proposals/components/InvitationForm.svelte';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { proposals as proposalsText } from '#lib/i18n/proposals.js';

	const id = { ...directoryText, ...proposalsText };

	const session = getAccountSession();
	const membership = $derived(session.account?.membership);
	const targetId = $derived(page.url.searchParams.get('sasaran') ?? '');
	const productId = $derived(page.url.searchParams.get('produk') ?? '');
	const needId = $derived(page.url.searchParams.get('kebutuhan') ?? '');

	function open(proposal: Proposal) {
		void goto(`/app/ajuan/${proposal.id}`);
	}
</script>

<svelte:head>
	<title>{targetId === '' ? id.proposals.newTitle : id.directory.propose}</title>
</svelte:head>

<section class="trade-panel mx-auto flex w-full max-w-form flex-col gap-6">
	{#if membership && targetId !== ''}
		<DirectProposalView
			{targetId}
			{needId}
			preselectedProductId={productId}
			isConsignor={membership.organization_type === 'CONSIGNOR'}
			onsent={open}
		/>
	{:else}
		<PageHeading icon={FileSignature}>{id.proposals.newTitle}</PageHeading>
		<p class="text-base text-text-muted">{id.proposals.newIntro}</p>
		{#if membership}
			<InvitationForm isConsignor={membership.organization_type === 'CONSIGNOR'} onsent={open} />
		{/if}
	{/if}
</section>
