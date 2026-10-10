<script lang="ts">
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
		<h1 class="text-xl font-semibold text-text">{id.proposals.listTitle}</h1>
		<a
			href="/app/ajuan/baru"
			class="inline-flex min-h-12 items-center rounded-control bg-primary px-5 font-semibold text-on-accent"
		>
			{id.proposals.invite}
		</a>
	</div>
	<ProposalList {organizationId} />
</section>
