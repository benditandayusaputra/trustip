<script lang="ts">
	import type { Snippet } from 'svelte';
	import ClipboardList from '@lucide/svelte/icons/clipboard-list';
	import Alert from '#lib/components/ui/Alert.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import Tabs from '#lib/components/ui/Tabs.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { Partnership } from '#lib/api/partnerships.js';
	import { id } from '#lib/i18n/id.js';
	import { PartnerRecord } from '../state.svelte';
	import PartnerCatalog from './PartnerCatalog.svelte';
	import PartnerEnding from './PartnerEnding.svelte';
	import PublicListingToggle from './PublicListingToggle.svelte';
	import PartnerSummary from './PartnerSummary.svelte';

	interface Props {
		partnershipId: string;
		canAmend: boolean;
		isConsignor: boolean;
		agreement: Snippet;
		drops?: Snippet;
		stock?: Snippet;
		counts?: Snippet;
		settlements?: Snippet;
	}

	let {
		partnershipId,
		canAmend,
		isConsignor,
		agreement,
		drops,
		stock,
		counts,
		settlements
	}: Props = $props();

	const amendHref = $derived(`/app/mitra/${partnershipId}/perubahan`);

	const text = id.partners;
	const record = new PartnerRecord(apiClient);
	const tabs = Object.entries(text.tabs).map(([key, label]) => ({ key, label }));
	const upcoming: Record<string, string> = {
		stock: text.comingStock,
		drops: text.comingDrops,
		counts: text.comingCounts,
		settlements: text.comingSettlements
	};
	let selected = $state<string>('summary');

	$effect(() => {
		void record.load(partnershipId);
	});

	function replacePartnership(changed: Partnership) {
		record.replace(changed);
	}
</script>

{#if record.status === 'loading'}
	<Skeleton />
{:else if record.status === 'failed' || !record.partnership}
	<Alert title={id.partners.loadFailed}>{record.failure}</Alert>
{:else}
	{@const partnership = record.partnership}
	<h1 class="text-xl font-semibold text-text">{partnership.counterpart.name}</h1>
	{#if !isConsignor && partnership.status === 'ACTIVE'}
		<div class="flex flex-wrap gap-4">
			<a href="/app/titipan/baru?mitra={partnershipId}" class="font-semibold text-accent">
				{id.drops.recordDrop}
			</a>
			<a href="/app/hitung/baru?mitra={partnershipId}" class="font-semibold text-accent">
				{id.counts.count}
			</a>
		</div>
	{/if}
	<Tabs label={text.tabsLabel} {tabs} bind:selected>
		{#snippet panel(key: string)}
			{#if key === 'summary'}
				<div class="flex flex-col gap-6">
					<PartnerSummary {partnership} />
					{#if canAmend && partnership.status === 'ACTIVE'}
						{#if !isConsignor}
							<PublicListingToggle {partnership} onchanged={replacePartnership} />
						{/if}
						<PartnerEnding {partnership} onchanged={replacePartnership} />
					{/if}
				</div>
			{:else if key === 'agreement'}
				{@const isAmendable = canAmend && partnership.status === 'ACTIVE'}
				<div class="flex flex-col gap-6">
					{#if isAmendable}
						<a href={amendHref} class="font-semibold text-accent">{id.agreement.amend}</a>
					{/if}
					{@render agreement()}
					<PartnerCatalog
						items={record.catalog}
						proposeHref={isAmendable && isConsignor
							? (productId: string) => `${amendHref}?tambah=${productId}`
							: undefined}
					/>
				</div>
			{:else if key === 'stock' && stock}
				{@render stock()}
			{:else if key === 'counts' && counts}
				{@render counts()}
			{:else if key === 'drops' && drops}
				{@render drops()}
			{:else if key === 'settlements' && settlements}
				{@render settlements()}
			{:else}
				<EmptyState icon={ClipboardList} message={upcoming[key] ?? ''}>
					{#snippet action()}
						<a href="/app/mitra" class="font-semibold text-accent">{id.nav.partners}</a>
					{/snippet}
				</EmptyState>
			{/if}
		{/snippet}
	</Tabs>
{/if}
