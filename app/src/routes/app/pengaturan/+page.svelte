<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Settings from '@lucide/svelte/icons/settings';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import OrganizationSettings from '#lib/features/organization/components/OrganizationSettings.svelte';
	import SessionEnd from '#lib/features/account/components/SessionEnd.svelte';
	import { organization as id } from '#lib/i18n/organization.js';

	const session = getAccountSession();
	const membership = $derived(session.account?.membership);
</script>

<svelte:head>
	<title>{id.settings.title}</title>
</svelte:head>

<section class="flex w-full flex-col gap-6">
	<PageHeading icon={Settings}>{id.settings.title}</PageHeading>
	{#if membership}
		<OrganizationSettings
			organizationId={membership.organization_id}
			canEdit={membership.role === 'OWNER'}
			view="preferences"
		/>
	{/if}
	<div class="self-start"><SessionEnd everywhere /></div>
</section>
