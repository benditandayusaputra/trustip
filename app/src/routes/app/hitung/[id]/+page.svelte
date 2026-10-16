<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Calculator from '@lucide/svelte/icons/calculator';
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
	<PageHeading backHref="/app/hitung" icon={Calculator}>{id.counts.recapTitle}</PageHeading>
	{#key page.params.id}
		<CountView countId={page.params.id ?? ''} viewerOrganizationId={organizationId} />
	{/key}
</section>
