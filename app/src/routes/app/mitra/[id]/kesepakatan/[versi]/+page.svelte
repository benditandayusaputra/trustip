<script lang="ts">
	import { page } from '$app/state';
	import AgreementView from '#lib/features/agreements/components/AgreementView.svelte';
	import { id } from '#lib/i18n/id.js';

	const version = $derived(Number(page.params.versi ?? '0'));
</script>

<svelte:head>
	<title>{id.agreement.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.agreement.title}</h1>
	{#key `${page.params.id ?? ''}-${String(version)}`}
		<AgreementView partnershipId={page.params.id ?? ''} {version} />
	{/key}
</section>
