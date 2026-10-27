<script lang="ts">
	import Wallet from '@lucide/svelte/icons/wallet';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import CalendarDays from '@lucide/svelte/icons/calendar-days';
	import type { SettlementFilter, SettlementNarrowing } from '#lib/api/settlements.js';
	import { formatBusinessDate, formatRupiah } from '#lib/i18n/format.js';
	import { settlements as id } from '#lib/i18n/settlements.js';
	import { SettlementList, emptyFilter } from '../list.svelte';
	import { standingLabel, standingTone } from '../rows';

	interface Props {
		partnershipId?: string;
		narrowing?: SettlementNarrowing;
		heading?: string;
	}

	let { partnershipId = '', narrowing, heading }: Props = $props();

	const text = id.settlements;
	const list = new SettlementList(apiClient);
	let chosen = $derived<SettlementFilter>(emptyFilter(partnershipId));
	const filter = $derived<SettlementFilter>(
		narrowing ? { partnershipId, status: narrowing.status, age: narrowing.age } : chosen
	);
	const query = $derived(narrowing?.query.trim().toLocaleLowerCase('id') ?? '');
	const shown = $derived(
		list.items.filter(
			(settlement) =>
				settlement.number.toLocaleLowerCase('id').includes(query) ||
				settlement.counterpart.name.toLocaleLowerCase('id').includes(query)
		)
	);
	const statusOptions = [
		{ value: '', label: text.allStatuses },
		{ value: 'ISSUED', label: text.statuses.ISSUED },
		{ value: 'PAID', label: text.statuses.PAID }
	];
	const ageOptions = [
		{ value: '', label: text.allAges },
		...Object.entries(text.ages).map(([value, label]) => ({ value, label }))
	];
	const isFiltered = $derived(filter.status !== '' || filter.age !== '' || query !== '');

	$effect(() => {
		void list.load(filter);
	});
</script>

{#if heading}
	<div class="flex flex-wrap items-baseline justify-between gap-2">
		<h2 class="text-lg font-semibold text-text">{heading}</h2>
		{#if list.status === 'ready' && shown.length > 0}
			<p class="text-xs text-text-muted">{text.showing(shown.length)}</p>
		{/if}
	</div>
{/if}
{#if !narrowing}
	<fieldset class="filter-toolbar grid grid-cols-1 gap-4 md:grid-cols-2">
		<legend class="sr-only">{text.filters}</legend>
		<Select
			label={text.status}
			options={statusOptions}
			value={chosen.status}
			onchange={(status: string) =>
				(chosen = { ...chosen, status: status as SettlementFilter['status'] })}
		/>
		<Select
			label={text.age}
			options={ageOptions}
			value={chosen.age}
			onchange={(age: string) => (chosen = { ...chosen, age: age as SettlementFilter['age'] })}
		/>
	</fieldset>
{/if}
{#if list.status === 'loading'}
	<ContentSkeleton variant="list" />
{:else if list.status === 'failed'}
	<Alert title={text.listFailed}>{list.failure}</Alert>
{:else if shown.length === 0}
	<EmptyState icon={Wallet} message={isFiltered ? text.noMatch : text.empty}>
		{#snippet action()}
			<a href="/app/hitung" class="font-semibold text-accent">{text.toCounts}</a>
		{/snippet}
	</EmptyState>
{:else}
	<ul class="flex flex-col gap-2">
		{#each shown as settlement (settlement.id)}
			{@const late = settlement.standing.age !== 'PAID' && settlement.standing.is_overdue}
			<li>
				<a
					href="/app/setoran/{settlement.id}"
					class="surface-card flex items-center justify-between gap-4 p-5 hover:border-accent hover:bg-surface-hover {late
						? 'border-danger bg-danger-surface'
						: ''}"
				>
					<span class="flex flex-col gap-1">
						<span class="font-mono text-xs text-text-muted">{settlement.number}</span>
						<span class="text-lg font-semibold text-text">{settlement.counterpart.name}</span>
						<span
							class="flex items-center gap-2 text-sm {late ? 'text-danger' : 'text-text-muted'}"
						>
							<CalendarDays class="size-4 shrink-0" aria-hidden="true" />
							{text.dueOn(formatBusinessDate(settlement.due_date))}
						</span>
						<span class="text-sm text-text">
							{text.mustPay}:
							<span class="font-bold tabular-nums {late ? 'text-danger' : ''}"
								>{formatRupiah(settlement.consignor_share)}</span
							>
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
