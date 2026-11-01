<script lang="ts">
	import type { Snippet } from 'svelte';
	import ChevronLeft from '@lucide/svelte/icons/chevron-left';
	import SkipLink from './SkipLink.svelte';
	import NavigationContent from './NavigationContent.svelte';
	import PublicFooter from './PublicFooter.svelte';
	import PublicHeader from './PublicHeader.svelte';
	import { common as id } from '#lib/i18n/common.js';

	interface Props {
		children: Snippet;
		footerNote?: Snippet;
		pendingPath?: string | undefined;
		currentPath?: string;
		backHref?: string | undefined;
	}

	let { children, footerNote, pendingPath, currentPath, backHref }: Props = $props();

	function goBack(event: MouseEvent) {
		if (history.length > 1) {
			event.preventDefault();
			history.back();
		}
	}
</script>

<SkipLink />
<div class="flex min-h-dvh flex-col">
	<PublicHeader {currentPath} />
	<main id="konten" tabindex="-1" class="mx-auto w-full max-w-content flex-1 px-4 py-8 lg:px-8">
		{#if backHref}
			<a
				href={backHref}
				onclick={goBack}
				class="mb-4 inline-flex min-h-11 items-center gap-2 text-sm font-medium text-text-muted hover:text-accent"
				><ChevronLeft class="size-4" aria-hidden="true" />{id.layout.back}</a
			>
		{/if}
		<NavigationContent {pendingPath}>{@render children()}</NavigationContent>
	</main>
	<PublicFooter note={footerNote} />
</div>
