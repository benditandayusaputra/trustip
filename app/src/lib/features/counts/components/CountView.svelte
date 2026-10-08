<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import KpiCard from '#lib/components/ui/KpiCard.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import SealPanel from '#lib/components/ui/SealPanel.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatDateTime, formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { CountRecord } from '../record.svelte';
	import { lineRows } from '../rows';

	interface Props {
		countId: string;
		isStore: boolean;
	}

	let { countId, isStore }: Props = $props();

	const text = id.counts;
	const record = new CountRecord(apiClient);
	const columns = [
		{ key: 'product' as const, label: text.product },
		{ key: 'recorded' as const, label: text.recordedColumn, numeric: true },
		{ key: 'sold' as const, label: text.soldUnits, numeric: true },
		{ key: 'returned' as const, label: text.returnedUnits, numeric: true },
		{ key: 'kept' as const, label: text.kept, numeric: true },
		{ key: 'gross' as const, label: text.gross, numeric: true }
	];

	$effect(() => {
		void record.load(countId);
	});
</script>

{#if record.status === 'loading'}
	<Skeleton />
{:else if record.status === 'failed' || !record.count}
	<Alert title={text.recapFailed}>{record.failure}</Alert>
{:else}
	{@const count = record.count}
	<div class="flex flex-wrap items-center gap-2">
		<p class="font-mono text-sm text-text-muted">{count.number}</p>
		<Badge
			tone={count.status === 'CONFIRMED' ? 'success' : 'pending'}
			label={text.statuses[count.status]}
		/>
	</div>
	<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
		<KpiCard label={text.gross} value={formatRupiah(count.split.gross)} />
		<KpiCard label={text.storeShare} value={formatRupiah(count.split.store_share)} />
		<KpiCard label={text.consignorShare} value={formatRupiah(count.split.consignor_share)} />
	</div>
	<dl class="grid grid-cols-2 gap-2 text-sm">
		<dt class="text-text-muted">{text.store}</dt>
		<dd>{count.store.name}</dd>
		<dt class="text-text-muted">{text.consignor}</dt>
		<dd>{count.consignor.name}</dd>
		<dt class="text-text-muted">{text.countedAt}</dt>
		<dd>{formatDateTime(count.counted_at)} · {text.occasions[count.occasion]}</dd>
		<dt class="text-text-muted">{text.recordedBy}</dt>
		<dd>{count.recorded_by_name}</dd>
		<dt class="text-text-muted">{text.soldUnits}</dt>
		<dd class="font-mono tabular-nums">{formatInteger(count.sold_qty)}</dd>
		<dt class="text-text-muted">{text.returnedUnits}</dt>
		<dd class="font-mono tabular-nums">{formatInteger(count.returned_qty)}</dd>
	</dl>
	{#if count.note}
		<p class="text-sm text-text">{text.note}: {count.note}</p>
	{/if}
	<h2 class="text-lg font-semibold text-text">{text.lines}</h2>
	<ResponsiveTable caption={text.lines} {columns} rows={lineRows(count)} rowKey="id" />
	<SealPanel isValid={count.seal.is_valid} events={count.seal.events} />
	{#if isStore}
		<a href="/app/titipan/baru?mitra={count.partnership_id}" class="font-semibold text-accent">
			{text.newDrop}
		</a>
	{/if}
{/if}
