<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import DateField from '#lib/components/ui/DateField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { SettlementNarrowing } from '#lib/api/settlements.js';
	import type { RangePreset } from '#lib/domain/ranges.js';
	import { formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { reports as id } from '#lib/i18n/reports.js';
	import { settlements as settlementText } from '#lib/i18n/settlements.js';
	import { SettlementReportState } from '../recaps.svelte';
	import { settlementKinds as kinds } from '../rows';
	import SettlementPartnerTable from './SettlementPartnerTable.svelte';

	let {
		narrowing = $bindable({ status: '', age: '', query: '' })
	}: { narrowing?: SettlementNarrowing } = $props();

	const text = id.reports;
	const filterText = settlementText.settlements;
	const state = new SettlementReportState(apiClient, () => new Date());
	const presetOptions = Object.entries(text.presets).map(([value, label]) => ({ value, label }));
	const statusOptions = [
		{ value: '', label: filterText.allStatuses },
		{ value: 'ISSUED', label: filterText.statuses.ISSUED },
		{ value: 'PAID', label: filterText.statuses.PAID }
	];
	const ageOptions = [
		{ value: '', label: filterText.allAges },
		...Object.entries(filterText.ages).map(([value, label]) => ({ value, label }))
	];
	const rows = $derived(
		(state.report?.rows ?? []).filter((row) =>
			row.name.toLocaleLowerCase('id').includes(narrowing.query.trim().toLocaleLowerCase('id'))
		)
	);

	$effect(() => {
		void state.load(state.range);
	});
</script>

{#if state.status === 'loading'}
	<ContentSkeleton variant="report" />
{:else if state.status === 'failed' || !state.report}
	<Alert title={text.loadFailed}>{state.failure}</Alert>
{:else}
	{@const total = state.report.total}
	<div class="grid grid-cols-(--figure-columns) gap-4">
		{#each kinds as kind (kind.key)}
			{@const figure = total[kind.key]}
			{@const alarming = kind.key === 'overdue' && figure.count > 0}
			<div
				class="surface-card flex min-w-0 flex-col gap-2 p-5 {alarming
					? 'border-danger bg-danger-surface'
					: ''}"
			>
				<div class="flex items-center justify-between gap-2">
					<p
						class="text-xs font-semibold tracking-(--letter-spacing-label) uppercase {alarming
							? 'text-danger'
							: 'text-text-muted'}"
					>
						{kind.label}
					</p>
					{#if alarming}
						<span
							class="rounded-pill bg-surface px-2 text-xs font-semibold tracking-(--letter-spacing-label) text-danger uppercase"
							>{text.attention}</span
						>
					{:else}
						<span class="size-2 rounded-pill {kind.dot}" aria-hidden="true"></span>
					{/if}
				</div>
				<p class="flex flex-wrap items-baseline gap-2 font-bold tabular-nums">
					<span class="text-kpi-compact {kind.tone}">{formatInteger(figure.count)}</span>
					<span class="text-text-muted" aria-hidden="true">·</span>
					<span class="text-lg {kind.key === 'overdue' ? kind.tone : 'text-text'}"
						>{formatRupiah(figure.amount)}</span
					>
				</p>
				<p class="text-xs {alarming ? 'text-danger' : 'text-text-muted'}">
					{text.settlementCaptions[kind.key]}
				</p>
			</div>
		{/each}
	</div>
{/if}

<fieldset class="filter-toolbar grid grid-cols-(--figure-columns) gap-4 p-5">
	<legend class="sr-only">{text.filters}</legend>
	<Select
		label={filterText.status}
		options={statusOptions}
		value={narrowing.status}
		onchange={(chosen: string) =>
			(narrowing = { ...narrowing, status: chosen as SettlementNarrowing['status'] })}
	/>
	<Select
		label={filterText.age}
		options={ageOptions}
		value={narrowing.age}
		onchange={(chosen: string) =>
			(narrowing = { ...narrowing, age: chosen as SettlementNarrowing['age'] })}
	/>
	<Select
		label={text.range}
		options={presetOptions}
		value={state.preset}
		onchange={(chosen: string) => (state.preset = chosen as RangePreset)}
	/>
	<TextField label={text.search} type="search" bind:value={narrowing.query} />
	{#if state.preset === 'CUSTOM'}
		<DateField
			label={text.from}
			value={state.custom.from}
			onchange={(from: string) => (state.custom = { ...state.custom, from })}
		/>
		<DateField
			label={text.to}
			value={state.custom.to}
			onchange={(to: string) => (state.custom = { ...state.custom, to })}
		/>
	{/if}
</fieldset>

{#if state.status === 'ready' && state.report}
	<SettlementPartnerTable {rows} partnerCount={state.report.rows.length} />
{/if}
