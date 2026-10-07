<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		checked?: boolean;
		name?: string;
		error?: string | undefined;
		onchange?: (checked: boolean) => void;
		children: Snippet;
	}

	let { checked = $bindable(false), name, error, onchange, children }: Props = $props();

	const inputId = $props.id();
</script>

<div class="flex flex-col gap-2">
	<div class="flex min-h-11 items-start gap-4">
		<input
			id={inputId}
			type="checkbox"
			{name}
			bind:checked
			onchange={() => onchange?.(checked)}
			aria-invalid={error ? true : undefined}
			aria-describedby={error ? `${inputId}-error` : undefined}
			class="mt-1 size-6 shrink-0 accent-accent"
		/>
		<label for={inputId} class="text-base text-text">{@render children()}</label>
	</div>
	{#if error}
		<p id="{inputId}-error" class="text-sm text-danger">{error}</p>
	{/if}
</div>
