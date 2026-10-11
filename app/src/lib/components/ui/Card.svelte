<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		title?: string;
		fill?: boolean;
		actions?: Snippet;
		children: Snippet;
	}

	let { title, actions, children, fill = false }: Props = $props();
</script>

<section
	class="{fill
		? 'h-full'
		: ''} flex flex-col gap-4 rounded-card border border-border bg-surface p-4 compact:p-5"
>
	{#if title !== undefined || actions !== undefined}
		<header class="flex items-center justify-between gap-4">
			{#if title}
				<h2 class="text-lg">{title}</h2>
			{/if}
			{@render actions?.()}
		</header>
	{/if}
	{@render children()}
</section>
