<script lang="ts">
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
	<h1 class="text-xl font-semibold text-text">{id.settings.profileTitle}</h1>
	{#if membership}
		<OrganizationSettings
			organizationId={membership.organization_id}
			canEdit={membership.role === 'OWNER'}
		/>
	{/if}
</section>
