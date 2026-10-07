<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		inputId: string;
		label: string;
		hint?: string | undefined;
		error?: string | undefined;
		children: Snippet;
	}

	let { inputId, label, hint, error, children }: Props = $props();
</script>

<div class="flex flex-col gap-2">
	<label for={inputId} class="text-sm font-medium text-text">{label}</label>
	{@render children()}
	{#if hint}
		<p id="{inputId}-hint" class="text-xs text-text-muted">{hint}</p>
	{/if}
	{#if error}
		<p id="{inputId}-error" class="text-sm text-danger">{error}</p>
	{/if}
</div>
