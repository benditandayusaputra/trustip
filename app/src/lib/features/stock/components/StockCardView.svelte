<script lang="ts">
	import History from '@lucide/svelte/icons/history';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatDateTime, formatInteger } from '#lib/i18n/format.js';
	import { stock as id } from '#lib/i18n/stock.js';
	import { sourceHrefOf } from '../rows';
	import { StockCardRecord } from '../state.svelte';

	interface Props {
		partnershipId: string;
		productId: string;
	}

	let { partnershipId, productId }: Props = $props();

	const text = id.stock;
	const record = new StockCardRecord(apiClient);
	const key = $derived({ partnershipId, productId });
	const signed = (qty: number) => (qty > 0 ? `+${formatInteger(qty)}` : formatInteger(qty));

	$effect(() => {
		void record.load(key);
	});
</script>

{#if record.status === 'loading'}
	<ContentSkeleton variant="detail" />
{:else if record.status === 'failed' || !record.card}
	<Alert title={text.cardFailed}>{record.failure}</Alert>
{:else}
	{@const card = record.card}
	<p class="text-base text-text">{card.product_name} ({card.sku})</p>
	{#if card.entries.length === 0}
		<EmptyState icon={History} message={text.cardEmpty}>
			{#snippet action()}
				<a href="/app/mitra/{partnershipId}" class="font-semibold text-accent">{text.openPartner}</a
				>
			{/snippet}
		</EmptyState>
	{:else}
		<ol class="flex flex-col gap-2">
			{#each card.entries as entry (entry.id)}
				<li class="flex flex-col gap-1 rounded-control border border-border bg-surface p-4 text-sm">
					<span class="flex items-center justify-between gap-2">
						<span class="font-semibold text-text">{text.types[entry.type]}</span>
						<span class="font-mono font-semibold text-text tabular-nums">{signed(entry.qty)}</span>
					</span>
					<span class="text-text-muted"
						>{formatDateTime(entry.occurred_at)} · {entry.batch_code}</span
					>
					<span class="flex items-center justify-between gap-2">
						<a
							href={sourceHrefOf(entry.source_type, entry.source_id)}
							class="font-mono font-semibold text-accent"
						>
							{entry.source_number}
						</a>
						<span class="font-mono text-text-muted tabular-nums">
							{text.balanceAfter(formatInteger(entry.balance_after))}
						</span>
					</span>
				</li>
			{/each}
		</ol>
		{#if card.has_more}
			<Button variant="secondary" busy={record.busy} onclick={() => void record.loadMore(key)}>
				{text.loadMore}
			</Button>
		{/if}
	{/if}
{/if}
