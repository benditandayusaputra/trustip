<script lang="ts">
	import { page } from '$app/state';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import DropView from '#lib/features/drops/components/DropView.svelte';
	import { drops as id } from '#lib/i18n/drops.js';

	const session = getAccountSession();
	const organizationId = $derived(session.account?.membership?.organization_id ?? '');
</script>

<svelte:head>
	<title>{id.drops.detailTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.drops.detailTitle}</h1>
	{#key page.params.id}
		<DropView dropId={page.params.id ?? ''} viewerOrganizationId={organizationId} />
	{/key}
</section>
