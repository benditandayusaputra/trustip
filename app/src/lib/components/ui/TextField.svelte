<script lang="ts">
	import Search from '@lucide/svelte/icons/search';
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
		inputmode?: 'text' | 'numeric' | 'decimal' | 'tel' | 'search' | 'email' | 'url' | 'none';
		hint?: string | undefined;
		error?: string | undefined;
	}

	let {
		label,
		value = $bindable(''),
		name,
		type = 'text',
		autocomplete,
		required = false,
		maxlength,
		inputmode,
		hint,
		error
	}: Props = $props();

	const inputId = $props.id();
</script>

<FieldShell {inputId} {label} {hint} {error}>
	<div class="relative">
		{#if type === 'search'}<Search
				class="pointer-events-none absolute top-4 left-4 size-5 text-text-muted"
				aria-hidden="true"
			/>{/if}
		<input
			id={inputId}
			{name}
			{type}
			{autocomplete}
			{required}
			{maxlength}
			{inputmode}
			bind:value
			aria-invalid={error ? true : undefined}
			aria-describedby={describedBy(inputId, hint, error)}
			class="{inputClass} {type === 'search' ? 'pl-12' : ''}"
		/>
	</div>
</FieldShell>
