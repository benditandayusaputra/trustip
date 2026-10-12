<script lang="ts">
	import FileSignature from '@lucide/svelte/icons/file-signature';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import Tabs from '#lib/components/ui/Tabs.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { ProposalBox, ProposalSummary } from '#lib/api/proposals.js';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { proposals as id } from '#lib/i18n/proposals.js';
	import { ProposalInbox } from '../inbox.svelte';

	interface Props {
		organizationId: string;
	}

	let { organizationId }: Props = $props();

	const text = id.proposals;
	const inbox = new ProposalInbox(apiClient);
	const tabs = Object.entries(text.boxes).map(([key, label]) => ({ key, label }));
	let selected = $state<string>('INCOMING');

	$effect(() => {
		void inbox.load(selected === 'OUTGOING' ? 'OUTGOING' : 'INCOMING');
	});

	function turnOf(summary: ProposalSummary) {
		return summary.awaiting_org_id === organizationId ? text.yourTurn : text.theirTurn;
	}

	function emptyText(box: ProposalBox) {
		return box === 'OUTGOING' ? text.emptyOutgoing : text.emptyIncoming;
	}
</script>

<Tabs label={text.boxLabel} {tabs} bind:selected>
	{#snippet panel()}
		{#if inbox.status === 'loading'}
			<ContentSkeleton variant="list" />
		{:else if inbox.status === 'failed'}
			<Alert title={text.listFailed}>{inbox.failure}</Alert>
		{:else if inbox.items.length === 0}
			<EmptyState icon={FileSignature} message={emptyText(inbox.box)}>
				{#snippet action()}
					<a href="/app/ajuan/baru" class="font-semibold text-accent">{text.invite}</a>
				{/snippet}
			</EmptyState>
		{:else}
			<ul class="flex flex-col gap-2">
				{#each inbox.items as summary (summary.id)}
					<li>
						<a
							href="/app/ajuan/{summary.id}"
							class="flex flex-col gap-1 rounded-control border border-border bg-surface p-4 hover:bg-surface-hover"
						>
							<span class="flex flex-wrap items-center justify-between gap-2">
								<span class="text-base font-semibold text-text">
									{summary.counterpart_name || text.notJoined}
								</span>
								<Badge
									tone={summary.status === 'OPEN' ? 'pending' : 'neutral'}
									label={text.statuses[summary.status]}
								/>
							</span>
							<span class="text-sm text-text-muted">
								{summary.number} · {formatDateTime(summary.created_at)}
								{#if summary.status === 'OPEN'}
									· {turnOf(summary)}
								{/if}
							</span>
						</a>
					</li>
				{/each}
			</ul>
			{#if inbox.nextCursor}
				<Button variant="secondary" busy={inbox.busy} onclick={() => void inbox.loadMore()}>
					{text.loadMore}
				</Button>
			{/if}
		{/if}
	{/snippet}
</Tabs>
