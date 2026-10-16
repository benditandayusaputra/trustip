<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Settings from '@lucide/svelte/icons/settings';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import OrganizationSettings from '#lib/features/organization/components/OrganizationSettings.svelte';
	import { organization as id } from '#lib/i18n/organization.js';

	const session = getAccountSession();
	const membership = $derived(session.account?.membership);
</script>

<svelte:head>
	<title>{id.settings.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-form flex-col gap-6">
	<PageHeading icon={Settings}>{id.settings.profileTitle}</PageHeading>
	{#if membership}
		<OrganizationSettings
			organizationId={membership.organization_id}
			canEdit={membership.role === 'OWNER'}
		/>
	{/if}
</section>
