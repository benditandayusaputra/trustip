<script lang="ts">
	import type { Component, Snippet } from 'svelte';

	interface Props {
		title?: string;
		icon?: Component<{ class?: string; 'aria-hidden'?: 'true' }>;
		fill?: boolean;
		actions?: Snippet;
		children: Snippet;
	}

	let { title, icon: Icon, actions, children, fill = false }: Props = $props();
</script>

<section class="{fill ? 'h-full' : ''} surface-card flex flex-col gap-5 p-5 compact:p-6">
	{#if title !== undefined || actions !== undefined}
		<header class="card-heading">
			{#if title}
				<h2 class="flex items-center gap-2 text-lg">
					{#if Icon}<Icon class="size-5 shrink-0 text-accent" aria-hidden="true" />{/if}{title}
				</h2>
			{/if}
			{@render actions?.()}
		</header>
	{/if}
	{@render children()}
</section>
