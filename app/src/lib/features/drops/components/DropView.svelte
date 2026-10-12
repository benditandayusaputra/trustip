<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import OfflineMark from '#lib/components/ui/OfflineMark.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import SealPanel from '#lib/components/ui/SealPanel.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatDateTime, formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { drops as dropsText } from '#lib/i18n/drops.js';
	import { proposals as proposalsText } from '#lib/i18n/proposals.js';
	import { DropRecord } from '../record.svelte';
	import { itemRows, statusToneOf } from '../rows';
	import DropActions from './DropActions.svelte';

	const id = { ...dropsText, ...proposalsText };

	interface Props {
		dropId: string;
		viewerOrganizationId: string;
	}

	let { dropId, viewerOrganizationId }: Props = $props();

	const text = id.drops;
	const record = new DropRecord(apiClient);
	const columns = [
		{ key: 'product' as const, label: text.product },
		{ key: 'batch' as const, label: text.batch },
		{ key: 'expiry' as const, label: text.expiry },
		{ key: 'qty' as const, label: text.units, numeric: true },
		{ key: 'retail' as const, label: id.proposals.retailPrice, numeric: true }
	];

	$effect(() => {
		void record.load(dropId);
	});
</script>

{#if record.status === 'loading'}
	<ContentSkeleton variant="detail" />
{:else if record.status === 'failed' || !record.drop}
	<Alert title={text.detailFailed}>{record.failure}</Alert>
{:else}
	{@const drop = record.drop}
	<div class="flex flex-wrap items-center gap-2">
		<p class="font-mono text-sm text-text-muted">{drop.number}</p>
		<Badge tone={statusToneOf(drop.status)} label={text.statuses[drop.status]} />
	</div>
	{#if drop.recorded_offline}
		<OfflineMark recordedAt={drop.client_recorded_at} receivedAt={drop.received_at} />
	{/if}
	<dl class="grid grid-cols-2 gap-2 text-sm">
		<dt class="text-text-muted">{text.store}</dt>
		<dd>{drop.store.name}</dd>
		<dt class="text-text-muted">{text.consignor}</dt>
		<dd>{drop.consignor.name}</dd>
		<dt class="text-text-muted">{text.occurredAt}</dt>
		<dd>{formatDateTime(drop.occurred_at)}</dd>
		<dt class="text-text-muted">{text.recordedBy}</dt>
		<dd>{drop.recorded_by_name}</dd>
	</dl>
	{#if drop.cancellation}
		<p class="text-sm text-text">
			{text.cancelledBy(drop.cancellation.by_name, formatDateTime(drop.cancellation.at))}
		</p>
	{/if}
	<DropActions {record} {viewerOrganizationId} />
	{#if drop.confirmation}
		<p class="text-sm text-text">
			{drop.confirmation.is_automatic
				? text.autoConfirmed(formatDateTime(drop.confirmation.at))
				: text.confirmedBy(drop.confirmation.by_name, formatDateTime(drop.confirmation.at))}
		</p>
	{/if}
	<h2 class="text-lg font-semibold text-text">{text.items}</h2>
	<ResponsiveTable caption={text.items} {columns} rows={itemRows(drop)} rowKey="id" />
	<dl class="grid grid-cols-2 gap-2 text-sm">
		<dt class="text-text-muted">{text.units}</dt>
		<dd class="text-right font-mono tabular-nums">{formatInteger(drop.total_qty)}</dd>
		<dt class="text-text-muted">{text.gross}</dt>
		<dd class="text-right font-mono tabular-nums">{formatRupiah(drop.value_if_sold.gross)}</dd>
		<dt class="text-text-muted">{text.storeShare}</dt>
		<dd class="text-right font-mono tabular-nums">
			{formatRupiah(drop.value_if_sold.store_share)}
		</dd>
		<dt class="text-text-muted">{text.consignorShare}</dt>
		<dd class="text-right font-mono tabular-nums">
			{formatRupiah(drop.value_if_sold.consignor_share)}
		</dd>
	</dl>
	<SealPanel isValid={drop.seal.is_valid} events={drop.seal.events} />
{/if}
