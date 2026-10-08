<script lang="ts">
	import DateField from '#lib/components/ui/DateField.svelte';
	import NumberField from '#lib/components/ui/NumberField.svelte';
	import { id } from '#lib/i18n/id.js';
	import { isExpirySoon, maxQty, type DropErrors, type DropLine } from '../schema';

	interface Props {
		line: DropLine;
		dropDate: string;
		errors: DropErrors;
		onchange: (changes: Partial<DropLine>) => void;
	}

	let { line, dropDate, errors, onchange }: Props = $props();

	const text = id.drops;
	const isSoon = $derived(line.qty > 0 && isExpirySoon(line, dropDate));
</script>

<fieldset class="flex flex-col gap-4 rounded-card border border-border bg-surface p-4">
	<legend class="px-1 text-base font-semibold text-text">{line.name}</legend>
	<NumberField
		label={text.qty(line.unit)}
		bind:value={
			() => line.qty,
			(qty: number) => {
				onchange({ qty });
			}
		}
		max={maxQty}
		error={errors[`${line.productId}.qty`]}
	/>
	<DateField
		label={text.expiry}
		bind:value={
			() => line.expiryDate,
			(expiryDate: string) => {
				onchange({ expiryDate });
			}
		}
		min={dropDate}
		hint={isSoon ? text.expirySoon : text.expiryHint}
		error={errors[`${line.productId}.expiry`]}
	/>
</fieldset>
