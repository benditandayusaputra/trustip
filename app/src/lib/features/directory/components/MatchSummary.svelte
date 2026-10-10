<script lang="ts">
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { DirectoryStoreCard } from '#lib/api/directory.js';
	import { directory as id } from '#lib/i18n/directory.js';

	interface Props {
		match: NonNullable<DirectoryStoreCard['match']>;
		compact?: boolean;
	}

	let { match, compact = false }: Props = $props();

	const text = id.directory;
	const label = $derived(match.label ? text.labels[match.label] : undefined);
</script>

<p
	class="flex shrink-0 gap-2 {compact
		? 'items-center justify-between rounded-control bg-accent-surface px-4 py-2'
		: 'flex-col items-end'}"
>
	<span class="text-2xl font-mono font-bold text-accent" aria-label={text.score(match.score)}
		>{match.score}</span
	>
	{#if label}
		<Badge tone="info" {label} />
	{/if}
</p>
