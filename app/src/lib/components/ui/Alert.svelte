<script lang="ts">
	import type { Snippet } from 'svelte';
	import TriangleAlert from '@lucide/svelte/icons/triangle-alert';

	interface Props {
		title: string;
		children: Snippet;
		action?: Snippet;
	}

	let { title, children, action }: Props = $props();
</script>

<div
	class="flex flex-col gap-2 rounded-control border-l-4 border-l-danger bg-danger-surface p-4 text-text"
>
	<p class="flex items-center gap-2 font-semibold text-danger">
		<TriangleAlert class="size-5 shrink-0" aria-hidden="true" />
		{title}
	</p>
	<div class="text-base">{@render children()}</div>
	{#if action}
		<div>{@render action()}</div>
	{/if}
</div>
