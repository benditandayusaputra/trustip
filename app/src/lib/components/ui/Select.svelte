<script lang="ts">
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import FieldShell from './FieldShell.svelte';
	import { describedBy, inputClass } from './field';

	interface Option {
		value: string;
		label: string;
	}

	interface Props {
		label: string;
		options: Option[];
		value?: string;
		name?: string;
		required?: boolean;
		disabled?: boolean;
		hint?: string | undefined;
		error?: string | undefined;
		onchange?: (value: string) => void;
	}

	let {
		label,
		options,
		value = $bindable(''),
		name,
		required = false,
		disabled = false,
		hint,
		error,
		onchange
	}: Props = $props();

	const inputId = $props.id();
</script>

<FieldShell {inputId} {label} {hint} {error}>
	<div class="relative">
		<select
			id={inputId}
			{name}
			{required}
			{disabled}
			bind:value
			onchange={() => onchange?.(value)}
			aria-invalid={error ? true : undefined}
			aria-describedby={describedBy(inputId, hint, error)}
			class="{inputClass} appearance-none pr-12"
		>
			{#each options as option (option.value)}
				<option value={option.value}>{option.label}</option>
			{/each}
		</select>
		<ChevronDown
			class="pointer-events-none absolute inset-y-0 right-4 my-auto size-5 text-text-muted"
			aria-hidden="true"
		/>
	</div>
</FieldShell>
