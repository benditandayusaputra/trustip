<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import KpiCard from '#lib/components/ui/KpiCard.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import SealPanel from '#lib/components/ui/SealPanel.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatBusinessDate, formatDate, formatDateTime } from '#lib/i18n/format.js';
	import { formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { SettlementRecord } from '../record.svelte';
	import { lineRows, periodLabel, standingLabel, standingTone } from '../rows';

	interface Props {
		settlementId: string;
	}

	let { settlementId }: Props = $props();

	const text = id.settlements;
	const record = new SettlementRecord(apiClient);
	const columns = [
		{ key: 'product' as const, label: text.product },
		{ key: 'model' as const, label: text.model },
		{ key: 'qty' as const, label: text.qty, numeric: true },
		{ key: 'retail' as const, label: text.retailPrice, numeric: true },
		{ key: 'gross' as const, label: text.gross, numeric: true },
		{ key: 'storeShare' as const, label: text.storeShare, numeric: true },
		{ key: 'consignorShare' as const, label: text.consignorShare, numeric: true }
	];

	$effect(() => {
		void record.load(settlementId);
	});
</script>

{#if record.status === 'loading'}
	<Skeleton />
{:else if record.status === 'failed' || !record.settlement}
	<Alert title={text.loadFailed}>{record.failure}</Alert>
{:else}
	{@const settlement = record.settlement}
	<div class="flex flex-wrap items-center gap-2">
		<p class="font-mono text-sm text-text-muted">{settlement.number}</p>
		<Badge tone={standingTone(settlement.standing)} label={standingLabel(settlement.standing)} />
	</div>
	<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
		<KpiCard label={text.gross} value={formatRupiah(settlement.split.gross)} />
		<KpiCard label={text.storeShare} value={formatRupiah(settlement.split.store_share)} />
		<KpiCard label={text.mustPay} value={formatRupiah(settlement.split.consignor_share)} />
	</div>
	<dl class="grid grid-cols-2 gap-2 text-sm">
		<dt class="text-text-muted">{text.store}</dt>
		<dd>{settlement.store.name}</dd>
		<dt class="text-text-muted">{text.consignor}</dt>
		<dd>{settlement.consignor.name}</dd>
		<dt class="text-text-muted">{text.schedule}</dt>
		<dd>{text.schedules[settlement.schedule]}</dd>
		<dt class="text-text-muted">{text.period}</dt>
		<dd>{periodLabel(settlement.period)}</dd>
		<dt class="text-text-muted">{text.issuedAt}</dt>
		<dd>{formatDateTime(settlement.issued_at)}</dd>
		<dt class="text-text-muted">{text.soldUnits}</dt>
		<dd class="font-mono tabular-nums">{formatInteger(settlement.sold_qty)}</dd>
	</dl>
	<p class="text-sm text-text">
		{settlement.paid_at
			? text.paidOn(formatDate(settlement.paid_at))
			: text.dueOn(formatBusinessDate(settlement.due_date))}
	</p>
	<h2 class="text-lg font-semibold text-text">{text.lines}</h2>
	<ResponsiveTable caption={text.lines} {columns} rows={lineRows(settlement)} rowKey="key" />
	<h2 class="text-lg font-semibold text-text">{text.sources}</h2>
	<ul class="flex flex-col gap-2">
		{#each settlement.counts as count (count.id)}
			<li>
				<a href="/app/hitung/{count.id}" class="font-mono text-accent">{count.number}</a>
				<span class="text-sm text-text-muted">· {formatDateTime(count.counted_at)}</span>
			</li>
		{/each}
	</ul>
	<SealPanel isValid={settlement.seal.is_valid} events={settlement.seal.events} />
{/if}
