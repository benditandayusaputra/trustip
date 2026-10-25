<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { agreements as agreementsText } from '#lib/i18n/agreements.js';
	import { products as productsText } from '#lib/i18n/products.js';
	import { proposals as proposalsText } from '#lib/i18n/proposals.js';
	import { AgreementRecord } from '../record.svelte';
	import { productRows } from '../rows';

	const id = { ...agreementsText, ...productsText, ...proposalsText };

	interface Props {
		partnershipId: string;
		version?: number;
	}

	let { partnershipId, version = 0 }: Props = $props();

	const text = id.agreement;
	const terms = id.proposals;
	const record = new AgreementRecord(apiClient);
	const columns = [
		{ key: 'product' as const, label: id.products.name },
		{ key: 'model' as const, label: terms.model },
		{ key: 'retail' as const, label: terms.retailPrice, numeric: true },
		{ key: 'consignor' as const, label: terms.perUnitConsignor, numeric: true },
		{ key: 'store' as const, label: terms.perUnitStore, numeric: true }
	];

	$effect(() => {
		void record.load(partnershipId, version);
	});
</script>

{#if record.status === 'loading'}
	<ContentSkeleton variant="detail" />
{:else if record.status === 'failed' || !record.agreement}
	<Alert title={text.loadFailed}>{record.failure}</Alert>
{:else}
	{@const agreement = record.agreement}
	{@const trail = agreement.trail}
	<p class="font-mono text-sm text-text-muted">
		{agreement.number} · {text.version(agreement.version)}
	</p>
	<dl class="document-details grid grid-cols-2 gap-2 text-sm">
		<dt class="text-text-muted">{terms.consignor}</dt>
		<dd>{agreement.consignor.name}</dd>
		<dt class="text-text-muted">{terms.store}</dt>
		<dd>{agreement.store.name}</dd>
		<dt class="text-text-muted">{terms.schedule}</dt>
		<dd>{terms.schedules[agreement.terms.settlement_schedule]}</dd>
		<dt class="text-text-muted">{terms.paymentTerms}</dt>
		<dd class="font-mono tabular-nums">{agreement.terms.payment_terms_days}</dd>
		<dt class="text-text-muted">{terms.rhythm}</dt>
		<dd>{terms.rhythms[agreement.terms.count_rhythm]}</dd>
		{#if agreement.terms.notes}
			<dt class="text-text-muted">{terms.notes}</dt>
			<dd>{agreement.terms.notes}</dd>
		{/if}
	</dl>
	<ul class="flex list-disc flex-col gap-1 pl-5 text-sm text-text-muted">
		{#each terms.fixedTermsItems as fixed (fixed)}
			<li>{fixed}</li>
		{/each}
	</ul>
	<h2 class="text-lg font-semibold text-text">{text.products}</h2>
	{#if agreement.items.length === 0}
		<p class="text-sm text-text-muted">{text.noProducts}</p>
	{:else}
		<ResponsiveTable caption={text.products} {columns} rows={productRows(agreement)} rowKey="id" />
	{/if}
	<h2 class="text-lg font-semibold text-text">{text.trail}</h2>
	<p class="text-sm text-text">
		{text.proposedBy(
			trail.proposed_by.name,
			trail.proposed_by.organization_name,
			trail.source_proposal_number
		)}
	</p>
	<p class="text-sm text-text">
		{text.acceptedBy(
			trail.accepted_by.name,
			trail.accepted_by.organization_name,
			formatDateTime(trail.accepted_at)
		)}
	</p>
	<a href="/app/ajuan/{trail.source_proposal_id}" class="text-sm font-semibold text-accent">
		{text.backToProposal}
	</a>
{/if}
