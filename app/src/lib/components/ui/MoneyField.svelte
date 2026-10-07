<script lang="ts">
	import FieldShell from './FieldShell.svelte';
	import { describedBy, inputClass } from './field';
	import { id } from '#lib/i18n/id.js';
	import { formatInteger, parseDigits } from '#lib/i18n/format.js';

	interface Props {
		label: string;
		value?: number | undefined;
		name?: string;
		hint?: string;
		error?: string;
	}

	let { label, value = $bindable(), name, hint, error }: Props = $props();

	const inputId = $props.id();
	const display = $derived(value === undefined ? '' : formatInteger(value));

	function readInput(event: Event & { currentTarget: HTMLInputElement }) {
		const input = event.currentTarget;
		value = parseDigits(input.value);
		input.value = display;
	}
</script>

<FieldShell {inputId} {label} {hint} {error}>
	<div class="relative">
		<span
			class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 font-mono text-text-muted"
			aria-hidden="true">{id.form.currencyPrefix}</span
		>
		<input
			id={inputId}
			{name}
			type="text"
			inputmode="numeric"
			autocomplete="off"
			value={display}
			oninput={readInput}
			aria-invalid={error ? true : undefined}
			aria-describedby={describedBy(inputId, hint, error)}
			class="{inputClass} pl-12 text-right font-mono font-medium tabular-nums"
		/>
	</div>
</FieldShell>
