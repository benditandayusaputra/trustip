<script lang="ts">
	import type { Snippet } from 'svelte';
	import PageSkeleton from './PageSkeleton.svelte';
	interface Props {
		pendingPath?: string | undefined;
		children: Snippet;
	}
	let { pendingPath, children }: Props = $props();
	const waiting = $derived(pendingPath !== undefined);
</script>

{#if pendingPath}
	<div data-navigation-loading><PageSkeleton path={pendingPath} /></div>
{/if}
<div
	class={waiting ? 'hidden' : 'contents'}
	aria-hidden={waiting ? 'true' : undefined}
	inert={waiting}
>
	{@render children()}
</div>
