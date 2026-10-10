<script lang="ts">
	import Calculator from '@lucide/svelte/icons/calculator';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import DateField from '#lib/components/ui/DateField.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { CountFilter } from '#lib/api/counts.js';
	import { formatDateTime, formatRupiah } from '#lib/i18n/format.js';
	import { counts as countsText } from '#lib/i18n/counts.js';
	import { drops as dropsText } from '#lib/i18n/drops.js';
	import { CountHistory, emptyFilter } from '../history.svelte';

	const id = { ...countsText, ...dropsText };

	interface Props {
		partnershipId?: string;
		countHref?: string | undefined;
	}

	let { partnershipId = '', countHref }: Props = $props();

	const text = id.counts;
	const history = new CountHistory(apiClient);
	let filter = $derived<CountFilter>(emptyFilter(partnershipId));
	const partnerOptions = $derived([
		{ value: '', label: id.drops.allPartners },
		...history.partners.map((partner) => ({ value: partner.id, label: partner.counterpart.name }))
	]);
	const isFiltered = $derived(filter.from !== '' || filter.to !== '');

	$effect(() => {
		if (partnershipId === '') {
			void history.loadPartners();
		}
	});

	$effect(() => {
		void history.load(filter);
	});
</script>

<fieldset class="grid grid-cols-1 gap-4 md:grid-cols-3">
	<legend class="sr-only">{text.filters}</legend>
	{#if partnershipId === ''}
		<Select
			label={id.drops.partner}
			options={partnerOptions}
			value={filter.partnershipId}
			onchange={(chosen: string) => (filter = { ...filter, partnershipId: chosen })}
		/>
	{/if}
	<DateField
		label={id.drops.from}
		value={filter.from}
		onchange={(from: string) => (filter = { ...filter, from })}
	/>
	<DateField
		label={id.drops.to}
		value={filter.to}
		onchange={(to: string) => (filter = { ...filter, to })}
	/>
</fieldset>
{#if history.status === 'loading'}
	<Skeleton />
{:else if history.status === 'failed'}
	<Alert title={text.historyFailed}>{history.failure}</Alert>
{:else if history.items.length === 0}
	<EmptyState icon={Calculator} message={isFiltered ? text.historyNoMatch : text.historyEmpty}>
		{#snippet action()}
			{#if countHref}
				<a href={countHref} class="font-semibold text-accent">{text.count}</a>
			{/if}
		{/snippet}
	</EmptyState>
{:else}
	<ul class="flex flex-col gap-2">
		{#each history.items as count (count.id)}
			<li>
				<a
					href="/app/hitung/{count.id}"
					class="flex items-center justify-between gap-4 rounded-control border border-border bg-surface p-4 hover:bg-surface-hover"
				>
					<span class="flex flex-col gap-1">
						<span class="font-mono text-sm text-text-muted">{count.number}</span>
						<span class="text-base font-semibold text-text">{count.counterpart.name}</span>
						<span class="text-sm text-text-muted">
							{formatDateTime(count.counted_at)} · {text.summary(
								count.sold_qty,
								count.returned_qty
							)}
						</span>
						<span class="font-mono text-sm text-text tabular-nums">
							{text.consignorShare}: {formatRupiah(count.consignor_share)}
						</span>
					</span>
					<Badge
						tone={count.status === 'CONFIRMED' ? 'success' : 'pending'}
						label={text.statuses[count.status]}
					/>
				</a>
			</li>
		{/each}
	</ul>
	{#if history.nextCursor}
		<Button variant="secondary" busy={history.busy} onclick={() => void history.loadMore()}>
			{text.loadMore}
		</Button>
	{/if}
{/if}
