<script lang="ts">
	import type { Snippet } from 'svelte';
	interface Props {
		count: number;
		previewCount?: number | undefined;
		moreLabel: string;
		lessLabel: string;
		children: Snippet<[number]>;
	}
	let {
		count,
		previewCount = Number.MAX_SAFE_INTEGER,
		moreLabel,
		lessLabel,
		children
	}: Props = $props();
	let expanded = $state(false);
	const contentId = $props.id();
	const limit = $derived(expanded ? count : previewCount);
</script>

<div class="flex flex-col gap-4">
	<div id={contentId}>{@render children(limit)}</div>
	{#if count > previewCount}
		<button
			type="button"
			aria-expanded={expanded}
			aria-controls={contentId}
			onclick={() => (expanded = !expanded)}
			class="min-h-11 self-start rounded-control border border-border-strong px-4 py-2 text-sm font-semibold text-accent hover:bg-surface-hover"
		>
			{expanded ? lessLabel : moreLabel}
		</button>
	{/if}
</div>
