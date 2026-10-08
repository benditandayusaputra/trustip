<script lang="ts">
	import Warehouse from '@lucide/svelte/icons/warehouse';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Card from '#lib/components/ui/Card.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatBusinessDate, formatInteger } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { flagToneOf, groupByProduct } from '../rows';
	import { StockHoldings } from '../state.svelte';

	interface Props {
		emptyHref: string;
		emptyLabel: string;
	}

	let { emptyHref, emptyLabel }: Props = $props();

	const text = id.stock;
	const holdings = new StockHoldings(apiClient);
	const groups = $derived(groupByProduct(holdings.items));

	$effect(() => {
		void holdings.load();
	});
</script>

{#if holdings.status === 'loading'}
	<Skeleton />
{:else if holdings.status === 'failed'}
	<Alert title={text.loadFailed}>{holdings.failure}</Alert>
{:else if groups.length === 0}
	<EmptyState icon={Warehouse} message={text.empty}>
		{#snippet action()}
			<a href={emptyHref} class="font-semibold text-accent">{emptyLabel}</a>
		{/snippet}
	</EmptyState>
{:else}
	{#each groups as group (group.productId)}
		<Card title={group.productName}>
			<p class="font-mono text-lg font-semibold text-text tabular-nums">
				{text.total(formatInteger(group.balance), group.unit)}
			</p>
			<ul class="flex flex-col gap-2">
				{#each group.holdings as holding (holding.partnership_id)}
					<li class="flex flex-wrap items-center justify-between gap-2 text-sm">
						<a href="/app/mitra/{holding.partnership_id}" class="font-semibold text-accent">
							{holding.counterpart.name}
						</a>
						<span class="flex items-center gap-2">
							<span class="font-mono text-text tabular-nums">{formatInteger(holding.balance)}</span>
							<Badge tone={flagToneOf(holding.flag)} label={text.flags[holding.flag]} />
						</span>
						{#if holding.nearest_expiry}
							<span class="w-full text-xs text-text-muted">
								{text.nearestExpiry(formatBusinessDate(holding.nearest_expiry))}
							</span>
						{/if}
					</li>
				{/each}
			</ul>
		</Card>
	{/each}
{/if}
