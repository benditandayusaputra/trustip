<script lang="ts">
	import Warehouse from '@lucide/svelte/icons/warehouse';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Card from '#lib/components/ui/Card.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatInteger } from '#lib/i18n/format.js';
	import { drops as dropsText } from '#lib/i18n/drops.js';
	import { stock as stockText } from '#lib/i18n/stock.js';
	import { batchRows, flagToneOf } from '../rows';
	import { PartnershipStock, suggestedQtyOf } from '../state.svelte';

	const id = { ...dropsText, ...stockText };

	interface Props {
		partnershipId: string;
		recordHref?: string | undefined;
	}

	let { partnershipId, recordHref }: Props = $props();

	const text = id.stock;
	const stock = new PartnershipStock(apiClient);
	const columns = [
		{ key: 'batch' as const, label: text.batch },
		{ key: 'expiry' as const, label: text.expiry },
		{ key: 'age' as const, label: text.age },
		{ key: 'remaining' as const, label: text.remaining, numeric: true },
		{ key: 'flag' as const, label: text.flag }
	];

	$effect(() => {
		void stock.load(partnershipId);
	});
</script>

{#if stock.status === 'loading'}
	<ContentSkeleton variant="list" />
{:else if stock.status === 'failed'}
	<Alert title={text.loadFailed}>{stock.failure}</Alert>
{:else if stock.products.length === 0}
	<EmptyState icon={Warehouse} message={text.empty}>
		{#snippet action()}
			{#if recordHref}
				<a href={recordHref} class="font-semibold text-accent">{id.drops.recordDrop}</a>
			{/if}
		{/snippet}
	</EmptyState>
{:else}
	{#each stock.products as product (product.product_id)}
		{@const suggestedQty = suggestedQtyOf(stock.suggestions, product.product_id)}
		<Card title={product.product_name}>
			<div class="flex flex-wrap items-center gap-2">
				<p class="font-mono text-lg font-semibold text-text tabular-nums">
					{text.total(formatInteger(product.balance), product.unit)}
				</p>
				<Badge tone={flagToneOf(product.flag)} label={text.flags[product.flag]} />
			</div>
			<p class="text-sm text-text-muted">
				{suggestedQty === undefined
					? id.drops.notEnoughData
					: id.drops.suggestedQty(suggestedQty, product.unit)}
			</p>
			<ResponsiveTable
				caption={text.batches(product.product_name)}
				{columns}
				rows={batchRows(product)}
				rowKey="id"
			/>
			<a
				href="/app/mitra/{partnershipId}/stok/{product.product_id}"
				class="text-sm font-semibold text-accent"
			>
				{text.openCard(product.product_name)}
			</a>
		</Card>
	{/each}
{/if}
