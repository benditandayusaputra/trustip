<script lang="ts">
	import type { Component, Snippet } from 'svelte';

	interface Props {
		icon: Component<{ class?: string; 'aria-hidden'?: 'true' }>;
		message: string;
		action: Snippet;
	}

	let { icon: Icon, message, action }: Props = $props();
</script>

<div class="flex flex-col items-center gap-4 px-4 py-8 text-center">
	<Icon class="size-12 text-text-muted" aria-hidden="true" />
	<p class="max-w-form text-base text-text">{message}</p>
	{@render action()}
</div>
