<script lang="ts">
	import Card from '#lib/components/ui/Card.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatDate, formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { settlements as id } from '#lib/i18n/settlements.js';
	import { RunningRecord } from '../list.svelte';

	interface Props {
		partnershipId: string;
	}

	let { partnershipId }: Props = $props();

	const text = id.settlements;
	const loadingLines = 4;
	const record = new RunningRecord(apiClient);

	$effect(() => {
		void record.load(partnershipId);
	});
</script>

<Card title={text.running}>
	<div class="flex flex-col gap-2">
		<p class="text-sm text-text-muted">{text.runningHint}</p>
		{#if record.status === 'loading'}
			<Skeleton lines={loadingLines} />
		{:else if record.status === 'failed' || !record.running}
			<p class="text-sm text-text">{text.runningFailed}</p>
		{:else if !record.running.has_sales}
			<p class="text-sm text-text">{text.runningEmpty}</p>
		{:else}
			{@const running = record.running}
			<dl class="grid grid-cols-2 gap-2 text-sm">
				<dt class="text-text-muted">{text.soldUnits}</dt>
				<dd class="font-mono tabular-nums">{formatInteger(running.sold_qty)}</dd>
				<dt class="text-text-muted">{text.gross}</dt>
				<dd class="font-mono tabular-nums">{formatRupiah(running.split.gross)}</dd>
				<dt class="text-text-muted">{text.storeShare}</dt>
				<dd class="font-mono tabular-nums">{formatRupiah(running.split.store_share)}</dd>
				<dt class="text-text-muted">{text.mustPay}</dt>
				<dd class="font-mono tabular-nums">{formatRupiah(running.split.consignor_share)}</dd>
			</dl>
			{#if running.since}
				<p class="text-sm text-text-muted">{text.runningSince(formatDate(running.since))}</p>
			{/if}
		{/if}
	</div>
</Card>
