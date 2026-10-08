<script lang="ts">
	import FieldShell from './FieldShell.svelte';
	import { describedBy, inputClass } from './field';

	interface Props {
		label: string;
		value?: string;
		min?: string;
		max?: string;
		name?: string;
		hint?: string | undefined;
		error?: string | undefined;
		onchange?: (value: string) => void;
	}

	let { label, value = $bindable(''), min, max, name, hint, error, onchange }: Props = $props();

	const inputId = $props.id();
</script>

<FieldShell {inputId} {label} {hint} {error}>
	<input
		id={inputId}
		{name}
		type="date"
		{min}
		{max}
		bind:value
		onchange={() => onchange?.(value)}
		aria-invalid={error ? true : undefined}
		aria-describedby={describedBy(inputId, hint, error)}
		class="{inputClass} font-mono tabular-nums"
	/>
</FieldShell>
