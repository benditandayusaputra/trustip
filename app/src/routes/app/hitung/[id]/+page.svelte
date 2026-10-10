<script lang="ts">
	import { page } from '$app/state';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import CountView from '#lib/features/counts/components/CountView.svelte';
	import { counts as id } from '#lib/i18n/counts.js';

	const session = getAccountSession();
	const organizationId = $derived(session.account?.membership?.organization_id ?? '');
</script>

<svelte:head>
	<title>{id.counts.recapTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.counts.recapTitle}</h1>
	{#key page.params.id}
		<CountView countId={page.params.id ?? ''} viewerOrganizationId={organizationId} />
	{/key}
</section>
