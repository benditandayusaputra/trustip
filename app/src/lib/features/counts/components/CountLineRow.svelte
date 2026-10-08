<script lang="ts">
	import NumberField from '#lib/components/ui/NumberField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import type { ReturnReason } from '#lib/domain/count.js';
	import { formatInteger } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import type { CountErrors, CountLineValues, InputMode, withChanges } from '../schema';

	interface Props {
		line: CountLineValues;
		mode: InputMode;
		errors: CountErrors;
		onchange: (changes: Parameters<typeof withChanges>[1]) => void;
	}

	let { line, mode, errors, onchange }: Props = $props();

	const text = id.counts;
	const reasonOptions = [
		{ value: '', label: text.chooseReason },
		...Object.entries(text.reasons).map(([value, label]) => ({ value, label }))
	];
	const soldQty = $derived(line.recordedQty - line.countedQty);
	const keptQty = $derived(line.countedQty - line.returnedQty);
	const error = $derived(errors[`${line.productId}.line`]);
</script>

<fieldset class="flex flex-col gap-4 rounded-card border border-border bg-surface p-4">
	<legend class="px-1 text-base font-semibold text-text">{line.name}</legend>
	<p class="font-mono text-sm text-text-muted tabular-nums">
		{text.recorded(formatInteger(line.recordedQty), line.unit)}
	</p>
	{#if mode === 'remaining'}
		<NumberField
			label={text.remaining(line.unit)}
			bind:value={
				() => line.countedQty,
				(countedQty: number) => {
					onchange({ countedQty });
				}
			}
			{error}
		/>
	{:else}
		<NumberField
			label={text.sold(line.unit)}
			max={line.recordedQty}
			bind:value={
				() => soldQty,
				(sold: number) => {
					onchange({ countedQty: line.recordedQty - sold });
				}
			}
			{error}
		/>
	{/if}
	<NumberField
		label={text.returned(line.unit)}
		bind:value={
			() => line.returnedQty,
			(returnedQty: number) => {
				onchange({ returnedQty });
			}
		}
	/>
	{#if line.returnedQty > 0}
		<Select
			label={text.reason}
			options={reasonOptions}
			value={line.returnReason}
			onchange={(reason: string) => {
				onchange({ returnReason: reason as ReturnReason | '' });
			}}
		/>
	{/if}
	<p class="text-sm text-text" aria-live="polite">{text.lineSummary(soldQty, keptQty)}</p>
</fieldset>
