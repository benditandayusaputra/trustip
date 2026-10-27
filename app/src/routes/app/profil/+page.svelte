<script lang="ts">
	import UserRound from '@lucide/svelte/icons/user-round';
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import AccountCard from '#lib/features/account/components/AccountCard.svelte';
	import OrganizationSettings from '#lib/features/organization/components/OrganizationSettings.svelte';
	import { organization as id } from '#lib/i18n/organization.js';

	const session = getAccountSession();
	const account = $derived(session.account);
</script>

<svelte:head>
	<title>{id.settings.profileTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-profile flex-col gap-6">
	<PageHeading icon={UserRound}>{id.settings.profileTitle}</PageHeading>
	{#if account}
		<AccountCard {account} />
		{#if account.membership}
			<OrganizationSettings
				organizationId={account.membership.organization_id}
				canEdit={account.membership.role === 'OWNER'}
				view="profile"
			/>
		{/if}
	{/if}
</section>
