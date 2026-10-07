<script lang="ts">
	import MoneyField from '#lib/components/ui/MoneyField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { exampleQuantity, splitOf } from '#lib/domain/pricing.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { termsOf, type InvitationErrors, type ItemValues } from '../schema';

	interface Props {
		item: ItemValues;
		index: number;
		errors: InvitationErrors;
	}

	let { item, index, errors }: Props = $props();

	const text = id.proposals;
	const modelOptions = Object.entries(text.models).map(([value, label]) => ({ value, label }));
	const example = $derived(splitOf(termsOf(item), exampleQuantity));
	const errorOf = (name: string) => errors[`items[${String(index)}].${name}`];
	const shown = (name: string) => {
		const message = errorOf(name);
		return message ? { error: message } : {};
	};
</script>

<fieldset class="flex flex-col gap-4 rounded-card border border-border bg-surface p-4">
	<legend class="px-1 text-base font-semibold text-text">{item.name}</legend>
	<Select label={text.model} options={modelOptions} bind:value={item.model} />
	<MoneyField label={text.retailPrice} bind:value={item.retailPrice} {...shown('retail_price')} />
	{#if item.model === 'FIXED_DEPOSIT'}
		<MoneyField
			label={text.depositPrice}
			bind:value={item.depositPrice}
			{...shown('deposit_price')}
		/>
	{:else}
		<TextField
			label={text.commission}
			hint={text.commissionHint}
			bind:value={item.commissionPercent}
			error={errorOf('commission_bps')}
		/>
	{/if}
	<dl class="grid grid-cols-2 gap-2 text-sm" aria-live="polite">
		<dt class="col-span-2 font-medium text-text-muted">{text.example(exampleQuantity)}</dt>
		<dt class="text-text-muted">{text.gross}</dt>
		<dd class="text-right font-mono font-medium tabular-nums">{formatRupiah(example.gross)}</dd>
		<dt class="text-text-muted">{text.storeShare}</dt>
		<dd class="text-right font-mono font-medium tabular-nums">
			{formatRupiah(example.storeShare)}
		</dd>
		<dt class="text-text-muted">{text.consignorShare}</dt>
		<dd class="text-right font-mono font-medium tabular-nums">
			{formatRupiah(example.consignorShare)}
		</dd>
	</dl>
</fieldset>
