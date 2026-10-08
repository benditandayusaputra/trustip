<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import type { Proposal } from '#lib/api/proposals.js';
	import InvitationForm from '#lib/features/proposals/components/InvitationForm.svelte';
	import InvitationLandingView from '#lib/features/proposals/components/InvitationLandingView.svelte';
	import ProposalDetail from '#lib/features/proposals/components/ProposalDetail.svelte';
	import ProposalList from '#lib/features/proposals/components/ProposalList.svelte';
	import AmendmentForm from '#lib/features/proposals/components/AmendmentForm.svelte';
	import DirectProposalView from '#lib/features/proposals/components/DirectProposalView.svelte';

	interface Props {
		view: 'form' | 'detail' | 'landing' | 'list' | 'amend' | 'direct';
		isConsignor?: boolean;
		proposalId?: string;
		viewerOrganizationId?: string;
		code?: string;
		needId?: string;
		ondone?: (proposal: Proposal) => void;
	}

	let {
		view,
		isConsignor = true,
		proposalId = '',
		viewerOrganizationId = '',
		code = '',
		needId = '',
		ondone = () => undefined
	}: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if view === 'form'}
	<InvitationForm {isConsignor} onsent={ondone} />
{:else if view === 'detail'}
	<ProposalDetail {proposalId} {viewerOrganizationId} />
{:else if view === 'list'}
	<ProposalList organizationId={viewerOrganizationId} />
{:else if view === 'direct'}
	<DirectProposalView
		targetId={proposalId}
		{needId}
		preselectedProductId={code}
		{isConsignor}
		onsent={ondone}
	/>
{:else if view === 'amend'}
	<AmendmentForm partnershipId={proposalId} prefillProductId={code} onsent={ondone} />
{:else}
	<InvitationLandingView
		{code}
		signInHref="/masuk?lanjut=%2Fundangan%2F{code}"
		registerHref="/daftar?lanjut=%2Fundangan%2F{code}"
		onboardingHref="/onboarding?lanjut=%2Fundangan%2F{code}"
		onclaimed={ondone}
	/>
{/if}
<Toaster />
