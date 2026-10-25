<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import OfflineMark from '#lib/components/ui/OfflineMark.svelte';
	import KpiCard from '#lib/components/ui/KpiCard.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatDateTime, formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { counts as countsText } from '#lib/i18n/counts.js';
	import { settlements as settlementsText } from '#lib/i18n/settlements.js';
	import { CountRecord } from '../record.svelte';
	import { lineRows } from '../rows';
	import CountConfirm from './CountConfirm.svelte';

	const id = { ...countsText, ...settlementsText };

	interface Props {
		countId: string;
		viewerOrganizationId: string;
	}

	let { countId, viewerOrganizationId }: Props = $props();

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
	<ContentSkeleton variant="detail" />
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
	{#if count.recorded_offline}
		<OfflineMark recordedAt={count.client_recorded_at} receivedAt={count.received_at} />
	{/if}
	<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
		<KpiCard label={text.gross} value={formatRupiah(count.split.gross)} />
		<KpiCard label={text.storeShare} value={formatRupiah(count.split.store_share)} />
		<KpiCard label={text.consignorShare} value={formatRupiah(count.split.consignor_share)} />
	</div>
	<dl class="document-details grid grid-cols-2 gap-2 text-sm">
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
	{#if count.confirmation}
		<p class="text-sm text-text">
			{count.confirmation.is_automatic
				? text.autoConfirmed(formatDateTime(count.confirmation.at))
				: text.confirmedBy(count.confirmation.by_name, formatDateTime(count.confirmation.at))}
		</p>
	{/if}
	<CountConfirm {record} {viewerOrganizationId} />
	{#if count.note}
		<p class="text-sm text-text">{text.note}: {count.note}</p>
	{/if}
	<h2 class="text-lg font-semibold text-text">{text.lines}</h2>
	<ResponsiveTable caption={text.lines} {columns} rows={lineRows(count)} rowKey="id" />
	<div class="flex flex-wrap gap-4">
		{#if count.settlement}
			<a href="/app/setoran/{count.settlement.id}" class="action-link">
				{count.store.id === viewerOrganizationId
					? id.settlements.recordPayment
					: id.settlements.openSettlement}
				<span class="font-mono">{count.settlement.number}</span>
			</a>
		{/if}
		{#if count.store.id === viewerOrganizationId}
			<a href="/app/titipan/baru?mitra={count.partnership_id}" class="secondary-link">
				{text.newDrop}
			</a>
		{/if}
	</div>
{/if}
