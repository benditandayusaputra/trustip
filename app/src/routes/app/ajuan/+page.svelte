<script lang="ts">
	import Plus from '@lucide/svelte/icons/plus';
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import FileSignature from '@lucide/svelte/icons/file-signature';
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import ProposalList from '#lib/features/proposals/components/ProposalList.svelte';
	import { proposals as id } from '#lib/i18n/proposals.js';

	const session = getAccountSession();
	const organizationId = $derived(session.account?.membership?.organization_id ?? '');
</script>

<svelte:head>
	<title>{id.proposals.listTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<PageHeading icon={FileSignature}>{id.proposals.listTitle}</PageHeading>
		<a href="/app/ajuan/baru" class="action-link">
			<Plus class="size-5" aria-hidden="true" />
			{id.proposals.invite}
		</a>
	</div>
	<ProposalList {organizationId} />
</section>
