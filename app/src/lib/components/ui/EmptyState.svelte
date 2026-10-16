<script lang="ts">
	import type { Component, Snippet } from 'svelte';

	interface Props {
		icon: Component<{ class?: string; 'aria-hidden'?: 'true' }>;
		message: string;
		action: Snippet;
	}

	let { icon: Icon, message, action }: Props = $props();
</script>

<div class="empty-state flex flex-col items-center gap-4 px-4 py-8 text-center">
	<span
		class="flex size-12 items-center justify-center rounded-control bg-accent-surface text-accent"
		><Icon class="size-6" aria-hidden="true" /></span
	>
	<p class="max-w-form text-base text-text">{message}</p>
	{@render action()}
</div>
