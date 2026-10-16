<script lang="ts">
	import type { Component, Snippet } from 'svelte';
	let {
		icon: Icon,
		children
	}: { icon: Component<{ class?: string; 'aria-hidden'?: 'true' }>; children: Snippet } = $props();
</script>

<div class="flex min-w-0 items-center gap-4">
	<span
		class="flex size-12 shrink-0 items-center justify-center rounded-card border border-border bg-surface text-accent shadow-card"
	>
		<Icon class="size-6" aria-hidden="true" />
	</span>
	<h1 class="page-heading min-w-0 font-semibold text-text">{@render children()}</h1>
</div>
