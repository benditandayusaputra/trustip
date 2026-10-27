<script lang="ts">
	import type { Component, Snippet } from 'svelte';
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import { common as text } from '#lib/i18n/common.js';
	import { getPageTitle } from '../layout/page-title.svelte';
	interface Props {
		icon: Component<{ class?: string; 'aria-hidden'?: 'true' }>;
		children: Snippet;
		backHref?: string;
	}
	let { icon: Icon, children, backHref }: Props = $props();
	const shell = getPageTitle();
	const inHeader = $derived(shell?.active === true);

	$effect(() => {
		if (!shell || !inHeader) {
			return;
		}
		return shell.show({ title: children, icon: Icon, backHref });
	});
</script>

{#if !inHeader}
	<div class="flex min-w-0 flex-col gap-2">
		{#if backHref}<a
				href={backHref}
				class="inline-flex min-h-11 items-center gap-2 self-start text-sm font-medium text-text-muted hover:text-accent"
				><ArrowLeft class="size-4" aria-hidden="true" />{text.layout.backToList}</a
			>{/if}
		<div class="flex min-w-0 items-center gap-2">
			<span
				class="flex size-11 shrink-0 items-center justify-center rounded-control border border-border bg-surface text-accent"
				><Icon class="size-5" aria-hidden="true" /></span
			>
			<h1 class="page-heading min-w-0 font-semibold text-text">{@render children()}</h1>
		</div>
	</div>
{/if}
