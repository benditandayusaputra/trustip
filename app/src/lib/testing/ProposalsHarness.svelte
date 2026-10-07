<script lang="ts">
	import Toaster from '#lib/components/ui/Toaster.svelte';
	import { Toaster as ToastQueue, setToaster } from '#lib/components/ui/toast.svelte.js';
	import type { Proposal } from '#lib/api/proposals.js';
	import InvitationForm from '#lib/features/proposals/components/InvitationForm.svelte';
	import InvitationLandingView from '#lib/features/proposals/components/InvitationLandingView.svelte';
	import ProposalDetail from '#lib/features/proposals/components/ProposalDetail.svelte';

	interface Props {
		view: 'form' | 'detail' | 'landing';
		isConsignor?: boolean;
		proposalId?: string;
		viewerOrganizationId?: string;
		code?: string;
		ondone?: (proposal: Proposal) => void;
	}

	let {
		view,
		isConsignor = true,
		proposalId = '',
		viewerOrganizationId = '',
		code = '',
		ondone = () => undefined
	}: Props = $props();

	setToaster(new ToastQueue());
</script>

{#if view === 'form'}
	<InvitationForm {isConsignor} onsent={ondone} />
{:else if view === 'detail'}
	<ProposalDetail {proposalId} {viewerOrganizationId} />
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
