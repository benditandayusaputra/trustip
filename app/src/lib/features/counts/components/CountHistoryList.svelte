<script lang="ts">
	import Calculator from '@lucide/svelte/icons/calculator';
	import ChefHat from '@lucide/svelte/icons/chef-hat';
	import Clock from '@lucide/svelte/icons/clock';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import DateField from '#lib/components/ui/DateField.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import StoreImage from '#lib/components/ui/StoreImage.svelte';
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
	<ContentSkeleton variant="list" />
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
			{@const counterpart = count.counterpart}
			<li>
				<a
					href="/app/hitung/{count.id}"
					class="surface-card flex flex-col gap-4 p-4 hover:border-accent hover:bg-surface-hover md:flex-row md:items-center md:gap-6"
				>
					<span class="flex min-w-0 flex-1 items-center gap-4">
						{#if counterpart.type === 'STORE' || counterpart.logo_url}
							<StoreImage url={counterpart.logo_url} class="size-12 rounded-control" />
						{:else}
							<span
								class="flex size-12 shrink-0 items-center justify-center rounded-control bg-accent-surface"
								><ChefHat class="size-6 text-accent" aria-hidden="true" /></span
							>
						{/if}
						<span class="flex min-w-0 flex-col gap-1">
							<span class="truncate text-base font-semibold text-text">{counterpart.name}</span>
							<span class="flex flex-wrap items-center gap-x-2 text-sm text-text-muted">
								<span class="font-mono">{count.number}</span>
								<span class="inline-flex items-center gap-1"
									><Clock class="size-4 shrink-0" aria-hidden="true" />{formatDateTime(
										count.counted_at
									)}</span
								>
							</span>
						</span>
					</span>
					<span
						class="grid grid-cols-3 items-center gap-x-6 gap-y-2 md:grid-cols-(--count-stats-columns)"
					>
						<span class="flex flex-col">
							<span class="text-xs text-text-muted">{text.soldUnits}</span>
							<span class="font-mono text-base font-semibold text-text tabular-nums"
								>{count.sold_qty}</span
							>
						</span>
						<span class="flex flex-col">
							<span class="text-xs text-text-muted">{text.returnedUnits}</span>
							<span class="font-mono text-base text-text tabular-nums">{count.returned_qty}</span>
						</span>
						<span class="flex flex-col">
							<span class="text-xs text-text-muted">{text.consignorShareShort}</span>
							<span class="font-mono text-base font-semibold text-text tabular-nums"
								>{formatRupiah(count.consignor_share)}</span
							>
						</span>
						<span class="col-span-3 md:col-span-1 md:justify-self-end">
							<Badge
								tone={count.status === 'CONFIRMED' ? 'success' : 'pending'}
								label={text.statuses[count.status]}
							/>
						</span>
					</span>
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
