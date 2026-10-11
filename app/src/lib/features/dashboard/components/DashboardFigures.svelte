<script lang="ts">
	import type { Dashboard } from '#lib/api/insights.js';
	import type { SalesReport } from '#lib/api/reports.js';
	import { formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { dashboard as id } from '#lib/i18n/dashboard.js';
	import type { Side } from '../cards';
	import DashboardKpi from './DashboardKpi.svelte';

	interface Props {
		figures: Dashboard['figures'];
		summary: SalesReport['summary'] | undefined;
		salesStatus: 'loading' | 'ready' | 'failed';
		isOwner: boolean;
		side: Side;
	}
	let { figures, summary, salesStatus, isOwner, side }: Props = $props();
	const text = id.dashboard;
	const loading = $derived(salesStatus === 'loading');
	const columns = $derived(side === 'STORE' ? 'md:grid-cols-3 xl:grid-cols-5' : 'lg:grid-cols-4');
</script>

<div class="grid grid-cols-2 items-stretch gap-4 {isOwner ? columns : 'md:grid-cols-3'}">
	{#if side === 'STORE'}
		<DashboardKpi
			metric="partners"
			label={text.activePartners}
			value={formatInteger(figures.active_partners)}
		/>
	{/if}
	<DashboardKpi
		metric="stock"
		label={side === 'STORE' ? text.unitsConsigned : text.unitsAtStores}
		value={formatInteger(figures.stock_units)}
	/>
	{#if isOwner}
		{#if side === 'STORE'}
			<DashboardKpi
				metric="gross"
				label={text.grossLast30}
				value={summary ? formatRupiah(summary.split.gross) : text.unavailable}
				{loading}
			/>
			<DashboardKpi
				metric="profit"
				tone="success"
				label={text.storeShareLast30}
				value={summary ? formatRupiah(summary.split.store_share) : text.unavailable}
				{loading}
			/>
		{:else}
			<DashboardKpi
				metric="sold"
				label={text.soldLast30}
				value={summary ? formatInteger(summary.sold_qty) : text.unavailable}
				{loading}
			/>
			<DashboardKpi
				metric="share"
				tone="success"
				label={text.consignorShareLast30}
				value={summary ? formatRupiah(summary.split.consignor_share) : text.unavailable}
				{loading}
			/>
		{/if}
	{/if}
	<DashboardKpi
		metric="payable"
		tone="pending"
		label={side === 'STORE' ? text.mustPayTotal : text.unpaidSettlements}
		value={formatRupiah(figures.unpaid.amount)}
		caption={side === 'STORE'
			? text.dueSoonCaption(figures.due_soon.count, formatRupiah(figures.due_soon.amount))
			: text.overdueCaption(figures.overdue.count, formatRupiah(figures.overdue.amount))}
	/>
</div>
{#if isOwner && salesStatus === 'failed'}
	<p role="status" class="text-sm text-danger">{text.salesFailed}</p>
{/if}
