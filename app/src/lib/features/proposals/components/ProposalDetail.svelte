<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { Proposal } from '#lib/api/proposals.js';
	import { formatBasisPoints, formatDateTime, formatRupiah } from '#lib/i18n/format.js';
	import { products as productsText } from '#lib/i18n/products.js';
	import { proposals as proposalsText } from '#lib/i18n/proposals.js';
	import { partyName } from '../links';
	import { ProposalRecord } from '../record.svelte';
	import { RevisionTrail } from '../trail.svelte';
	import CounterForm from './CounterForm.svelte';
	import InvitationShare from './InvitationShare.svelte';
	import ProposalActions from './ProposalActions.svelte';
	import RevisionHistory from './RevisionHistory.svelte';

	const id = { ...productsText, ...proposalsText };

	interface Props {
		proposalId: string;
		viewerOrganizationId: string;
	}

	let { proposalId, viewerOrganizationId }: Props = $props();

	const text = id.proposals;
	const record = new ProposalRecord(apiClient);
	const trail = new RevisionTrail(apiClient);
	let isCountering = $state(false);
	interface Column {
		key: keyof ReturnType<typeof rowsOf>[number];
		label: string;
		numeric?: boolean;
	}
	const columns: Column[] = [
		{ key: 'product', label: id.products.name },
		{ key: 'model', label: text.model },
		{ key: 'retail', label: text.retailPrice, numeric: true },
		{ key: 'consignor', label: text.perUnitConsignor, numeric: true },
		{ key: 'store', label: text.perUnitStore, numeric: true },
		{ key: 'change', label: text.change }
	];

	$effect(() => {
		void record.load(proposalId);
		void trail.load(proposalId);
	});

	function countered(proposal: Proposal) {
		record.proposal = proposal;
		isCountering = false;
		void trail.load(proposal.id);
	}

	function rowsOf(proposal: Proposal) {
		return proposal.revision.items.map((item) => ({
			id: item.product_id,
			product: `${item.product_name} (${item.sku})`,
			model:
				item.pricing_model === 'COMMISSION'
					? `${text.models.COMMISSION} ${formatBasisPoints(item.commission_bps)}`
					: text.models.FIXED_DEPOSIT,
			retail: formatRupiah(item.retail_price),
			consignor: formatRupiah(item.per_unit.consignor_share),
			store: formatRupiah(item.per_unit.store_share),
			change: changeLabel(item.product_id)
		}));
	}

	function changeLabel(productId: string): string {
		const change = trail.changeOf(productId);
		return change ? text.changes[change] : '';
	}

	function awaitingText(proposal: Proposal): string {
		if (proposal.awaiting_org_id === null) {
			return text.awaitingInvitee;
		}
		return text.awaiting(partyName(proposal, proposal.awaiting_org_id));
	}
</script>

{#if record.status === 'loading'}
	<Skeleton />
{:else if record.status === 'failed' || !record.proposal}
	<Alert title={text.loadFailed}>{record.failure}</Alert>
{:else}
	{@const proposal = record.proposal}
	{@const terms = proposal.revision.terms}
	<div class="flex flex-wrap items-center gap-2">
		<span class="font-mono text-sm text-text-muted">{proposal.number}</span>
		<Badge
			tone={proposal.status === 'OPEN' ? 'pending' : 'neutral'}
			label={text.statuses[proposal.status]}
		/>
	</div>
	{#if proposal.status === 'OPEN'}
		<p role="status" class="text-base text-text">{awaitingText(proposal)}</p>
		<ProposalActions
			{record}
			canDecide={proposal.awaiting_org_id === viewerOrganizationId}
			canWithdraw={proposal.revision.author_org_id === viewerOrganizationId}
			oncounter={() => (isCountering = true)}
		/>
		{#if isCountering}
			<CounterForm
				{proposal}
				counterpartName={partyName(proposal, proposal.revision.author_org_id)}
				oncountered={countered}
				oncancel={() => (isCountering = false)}
			/>
		{/if}
	{:else if proposal.status === 'DECLINED' && proposal.decision_reason}
		<p class="text-base text-text">{text.declineReason(proposal.decision_reason)}</p>
	{:else if proposal.partnership_id}
		<a href="/app/mitra/{proposal.partnership_id}" class="font-semibold text-accent">
			{text.openPartnership}
		</a>
	{/if}
	<dl class="grid grid-cols-2 gap-2 text-sm">
		<dt class="text-text-muted">{text.consignor}</dt>
		<dd>{proposal.consignor?.name ?? text.notJoined}</dd>
		<dt class="text-text-muted">{text.store}</dt>
		<dd>{proposal.store?.name ?? text.notJoined}</dd>
		<dt class="text-text-muted">{text.schedule}</dt>
		<dd>{text.schedules[terms.settlement_schedule]}</dd>
		<dt class="text-text-muted">{text.paymentTerms}</dt>
		<dd class="font-mono tabular-nums">{terms.payment_terms_days}</dd>
		<dt class="text-text-muted">{text.rhythm}</dt>
		<dd>{text.rhythms[terms.count_rhythm]}</dd>
		{#if terms.notes}
			<dt class="text-text-muted">{text.notes}</dt>
			<dd>{terms.notes}</dd>
		{/if}
		<dt class="text-text-muted">{id.proposals.status}</dt>
		<dd>{text.expiresAt(formatDateTime(proposal.expires_at))}</dd>
	</dl>
	{#if proposal.revision.items.length > 0}
		<ResponsiveTable caption={text.products} {columns} rows={rowsOf(proposal)} rowKey="id" />
	{/if}
	{#if trail.removed.length > 0}
		<p class="text-sm text-text-muted">
			{text.itemsRemoved(trail.removed.map((item) => item.product_name).join(', '))}
		</p>
	{/if}
	{#if proposal.revision.message}
		<blockquote class="border-l-4 border-l-border-strong pl-4 text-base text-text">
			{proposal.revision.message}
		</blockquote>
	{/if}
	<RevisionHistory records={trail.records} />
	{#if proposal.invite_code && proposal.awaiting_org_id === null && proposal.from_org_id === viewerOrganizationId}
		<InvitationShare {proposal} inviteCode={proposal.invite_code} />
	{/if}
{/if}
