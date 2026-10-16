<script lang="ts">
	import FieldShell from './FieldShell.svelte';
	import { describedBy, inputClass } from './field';
	interface Props {
		label: string;
		value?: string;
		name?: string;
		maxlength?: number;
		autocomplete?: HTMLTextAreaElement['autocomplete'];
		hint?: string | undefined;
		error?: string | undefined;
	}
	let {
		label,
		value = $bindable(''),
		name,
		maxlength,
		autocomplete,
		hint,
		error
	}: Props = $props();
	const inputId = $props.id();
</script>

<FieldShell {inputId} {label} {hint} {error}>
	<textarea
		id={inputId}
		{name}
		{maxlength}
		{autocomplete}
		bind:value
		aria-invalid={error ? true : undefined}
		aria-describedby={describedBy(inputId, hint, error)}
		class="{inputClass} min-h-(--textarea-height) resize-y py-4"></textarea>
</FieldShell>
