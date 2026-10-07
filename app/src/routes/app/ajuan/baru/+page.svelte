<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Proposal } from '#lib/api/proposals.js';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import InvitationForm from '#lib/features/proposals/components/InvitationForm.svelte';
	import { id } from '#lib/i18n/id.js';

	const session = getAccountSession();
	const membership = $derived(session.account?.membership);
</script>

<svelte:head>
	<title>{id.proposals.newTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.proposals.newTitle}</h1>
	<p class="text-base text-text-muted">{id.proposals.newIntro}</p>
	{#if membership}
		<InvitationForm
			isConsignor={membership.organization_type === 'CONSIGNOR'}
			onsent={(proposal: Proposal) => void goto(`/app/ajuan/${proposal.id}`)}
		/>
	{/if}
</section>
