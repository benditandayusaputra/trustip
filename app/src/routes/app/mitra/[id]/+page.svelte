<script lang="ts">
	import { page } from '$app/state';
	import AgreementView from '#lib/features/agreements/components/AgreementView.svelte';
	import PartnerPage from '#lib/features/partnerships/components/PartnerPage.svelte';
	import { id } from '#lib/i18n/id.js';

	const partnershipId = $derived(page.params.id ?? '');
</script>

<svelte:head>
	<title>{id.partners.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	{#key partnershipId}
		<PartnerPage {partnershipId}>
			{#snippet agreement()}
				<AgreementView {partnershipId} />
			{/snippet}
		</PartnerPage>
	{/key}
</section>
