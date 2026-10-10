<script lang="ts">
	import { page } from '$app/state';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import ProposalDetail from '#lib/features/proposals/components/ProposalDetail.svelte';
	import { proposals as id } from '#lib/i18n/proposals.js';

	const session = getAccountSession();
	const organizationId = $derived(session.account?.membership?.organization_id ?? '');
</script>

<svelte:head>
	<title>{id.proposals.detailTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.proposals.detailTitle}</h1>
	{#key page.params.id}
		<ProposalDetail proposalId={page.params.id ?? ''} viewerOrganizationId={organizationId} />
	{/key}
</section>
