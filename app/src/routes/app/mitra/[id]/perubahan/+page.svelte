<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Proposal } from '#lib/api/proposals.js';
	import AmendmentForm from '#lib/features/proposals/components/AmendmentForm.svelte';
	import { proposals as id } from '#lib/i18n/proposals.js';
</script>

<svelte:head>
	<title>{id.proposals.amendTitle}</title>
</svelte:head>

<section class="trade-panel mx-auto flex w-full max-w-form flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.proposals.amendTitle}</h1>
	<p class="text-base text-text-muted">{id.proposals.amendIntro}</p>
	{#key page.url.href}
		<AmendmentForm
			partnershipId={page.params.id ?? ''}
			prefillProductId={page.url.searchParams.get('tambah') ?? ''}
			onsent={(proposal: Proposal) => void goto(`/app/ajuan/${proposal.id}`)}
		/>
	{/key}
</section>
