<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { getDirectoryContact, type DirectoryContact } from '#lib/api/directory.js';
	import type { Proposal } from '#lib/api/proposals.js';
	import { id } from '#lib/i18n/id.js';
	import InvitationForm from './InvitationForm.svelte';

	interface Props {
		targetId: string;
		preselectedProductId: string;
		isConsignor: boolean;
		onsent: (proposal: Proposal) => void;
	}

	let { targetId, preselectedProductId, isConsignor, onsent }: Props = $props();

	const text = id.proposals;
	let contact = $state<DirectoryContact | undefined>();
	let status = $state<'loading' | 'ready' | 'missing'>('loading');

	$effect(() => {
		status = 'loading';
		getDirectoryContact(apiClient(), targetId)
			.then((loaded) => {
				contact = loaded;
				status = 'ready';
			})
			.catch(() => {
				status = 'missing';
			});
	});
</script>

{#if status === 'loading'}
	<p role="status" class="text-base text-text-muted">{text.targetLoading}</p>
{:else if status === 'missing' || !contact}
	<Alert title={text.targetMissing}>
		<a href="/app" class="font-semibold text-accent">{id.errorPage.backHome}</a>
	</Alert>
{:else}
	<h1 class="text-xl font-semibold text-text">{text.directTitle(contact.name)}</h1>
	<p class="text-base text-text-muted">{text.directIntro}</p>
	<InvitationForm
		{isConsignor}
		{targetId}
		{preselectedProductId}
		submitLabel={text.directSend}
		sentMessage={text.directSent}
		{onsent}
	/>
{/if}
