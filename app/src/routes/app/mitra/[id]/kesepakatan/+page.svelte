<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Handshake from '@lucide/svelte/icons/handshake';
	import { page } from '$app/state';
	import AgreementHistory from '#lib/features/agreements/components/AgreementHistory.svelte';
	import AgreementView from '#lib/features/agreements/components/AgreementView.svelte';
	import { agreements as id } from '#lib/i18n/agreements.js';
</script>

<svelte:head>
	<title>{id.agreement.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<PageHeading backHref="/app/mitra/{page.params.id}" icon={Handshake}
		>{id.agreement.title}</PageHeading
	>
	{#key page.params.id}
		<AgreementView partnershipId={page.params.id ?? ''} />
		<AgreementHistory partnershipId={page.params.id ?? ''} />
	{/key}
</section>
