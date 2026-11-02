<script lang="ts">
	import type { Snippet } from 'svelte';
	import PageSkeleton from './PageSkeleton.svelte';
	interface Props {
		pendingPath?: string | undefined;
		titled?: boolean;
		children: Snippet;
	}
	let { pendingPath, titled = true, children }: Props = $props();
	const skeletonDelayMs = 150;
	let shownPath = $state<string | undefined>();
	const waiting = $derived(shownPath !== undefined);

	$effect(() => {
		const path = pendingPath;
		if (path === undefined) {
			shownPath = undefined;
			return;
		}
		const timer = setTimeout(() => {
			shownPath = path;
		}, skeletonDelayMs);
		return () => {
			clearTimeout(timer);
		};
	});
</script>

{#if shownPath}
	<div data-navigation-loading><PageSkeleton path={shownPath} {titled} /></div>
{/if}
<div
	class={waiting ? 'hidden' : 'contents'}
	aria-hidden={waiting ? 'true' : undefined}
	inert={waiting}
>
	{@render children()}
</div>
