<script lang="ts">
	import Wallet from '@lucide/svelte/icons/wallet';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { SettlementFilter } from '#lib/api/settlements.js';
	import { formatBusinessDate, formatRupiah } from '#lib/i18n/format.js';
	import { settlements as id } from '#lib/i18n/settlements.js';
	import { SettlementList, emptyFilter } from '../list.svelte';
	import { standingLabel, standingTone } from '../rows';

	interface Props {
		partnershipId?: string;
	}

	let { partnershipId = '' }: Props = $props();

	const text = id.settlements;
	const list = new SettlementList(apiClient);
	let filter = $derived<SettlementFilter>(emptyFilter(partnershipId));
	const statusOptions = [
		{ value: '', label: text.allStatuses },
		{ value: 'ISSUED', label: text.statuses.ISSUED },
		{ value: 'PAID', label: text.statuses.PAID }
	];
	const ageOptions = [
		{ value: '', label: text.allAges },
		...Object.entries(text.ages).map(([value, label]) => ({ value, label }))
	];
	const isFiltered = $derived(filter.status !== '' || filter.age !== '');

	$effect(() => {
		void list.load(filter);
	});
</script>

<fieldset class="filter-toolbar grid grid-cols-1 gap-4 md:grid-cols-2">
	<legend class="sr-only">{text.filters}</legend>
	<Select
		label={text.status}
		options={statusOptions}
		value={filter.status}
		onchange={(chosen: string) =>
			(filter = { ...filter, status: chosen as SettlementFilter['status'] })}
	/>
	<Select
		label={text.age}
		options={ageOptions}
		value={filter.age}
		onchange={(chosen: string) => (filter = { ...filter, age: chosen as SettlementFilter['age'] })}
	/>
</fieldset>
{#if list.status === 'loading'}
	<ContentSkeleton variant="list" />
{:else if list.status === 'failed'}
	<Alert title={text.listFailed}>{list.failure}</Alert>
{:else if list.items.length === 0}
	<EmptyState icon={Wallet} message={isFiltered ? text.noMatch : text.empty}>
		{#snippet action()}
			<a href="/app/hitung" class="font-semibold text-accent">{text.toCounts}</a>
		{/snippet}
	</EmptyState>
{:else}
	<ul class="flex flex-col gap-2">
		{#each list.items as settlement (settlement.id)}
			<li>
				<a
					href="/app/setoran/{settlement.id}"
					class="surface-card flex items-center justify-between gap-4 p-5 hover:border-accent hover:bg-surface-hover"
				>
					<span class="flex flex-col gap-1">
						<span class="font-mono text-sm text-text-muted">{settlement.number}</span>
						<span class="text-base font-semibold text-text">{settlement.counterpart.name}</span>
						<span class="text-sm text-text-muted">
							{text.dueOn(formatBusinessDate(settlement.due_date))}
						</span>
						<span class="font-mono text-sm text-text tabular-nums">
							{text.mustPay}: {formatRupiah(settlement.consignor_share)}
						</span>
					</span>
					<Badge
						tone={standingTone(settlement.standing)}
						label={standingLabel(settlement.standing)}
					/>
				</a>
			</li>
		{/each}
	</ul>
	{#if list.nextCursor}
		<Button variant="secondary" busy={list.busy} onclick={() => void list.loadMore()}>
			{text.loadMore}
		</Button>
	{/if}
{/if}
