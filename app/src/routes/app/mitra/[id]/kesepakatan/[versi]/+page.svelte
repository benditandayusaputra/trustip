<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Handshake from '@lucide/svelte/icons/handshake';
	import { page } from '$app/state';
	import AgreementView from '#lib/features/agreements/components/AgreementView.svelte';
	import { agreements as id } from '#lib/i18n/agreements.js';

	const version = $derived(Number(page.params.versi ?? '0'));
</script>

<svelte:head>
	<title>{id.agreement.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<PageHeading icon={Handshake}>{id.agreement.title}</PageHeading>
	{#key `${page.params.id ?? ''}-${String(version)}`}
		<AgreementView partnershipId={page.params.id ?? ''} {version} />
	{/key}
</section>
