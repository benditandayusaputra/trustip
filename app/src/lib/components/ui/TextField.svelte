<script lang="ts">
	import FieldShell from './FieldShell.svelte';
	import { describedBy, inputClass } from './field';

	interface Props {
		label: string;
		value?: string;
		name?: string;
		type?: 'text' | 'email' | 'password' | 'tel' | 'search' | 'url';
		autocomplete?: HTMLInputElement['autocomplete'];
		required?: boolean;
		maxlength?: number;
		hint?: string;
		error?: string;
	}

	let {
		label,
		value = $bindable(''),
		name,
		type = 'text',
		autocomplete,
		required = false,
		maxlength,
		hint,
		error
	}: Props = $props();

	const inputId = $props.id();
</script>

<FieldShell {inputId} {label} {hint} {error}>
	<input
		id={inputId}
		{name}
		{type}
		{autocomplete}
		{required}
		{maxlength}
		bind:value
		aria-invalid={error ? true : undefined}
		aria-describedby={describedBy(inputId, hint, error)}
		class={inputClass}
	/>
</FieldShell>
