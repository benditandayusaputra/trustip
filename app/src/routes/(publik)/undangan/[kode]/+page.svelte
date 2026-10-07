<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Proposal } from '#lib/api/proposals.js';
	import { onboardingHref, registerHref, signInHref } from '#lib/features/auth/links.js';
	import InvitationLandingView from '#lib/features/proposals/components/InvitationLandingView.svelte';
	import { id } from '#lib/i18n/id.js';

	const code = $derived(page.params.kode ?? '');
	const here = $derived(`/undangan/${encodeURIComponent(code)}`);
</script>

<svelte:head>
	<title>{id.invitation.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6 py-8">
	<h1 class="text-xl font-semibold text-text">{id.invitation.title}</h1>
	{#key code}
		<InvitationLandingView
			{code}
			signInHref={signInHref(here)}
			registerHref={registerHref(here)}
			onboardingHref={onboardingHref(here)}
			onclaimed={(proposal: Proposal) => void goto(`/app/ajuan/${proposal.id}`)}
		/>
	{/key}
</section>
