<script lang="ts">
	import type { Snippet } from 'svelte';
	import SkipLink from './SkipLink.svelte';
	import NavigationContent from './NavigationContent.svelte';
	import PublicFooter from './PublicFooter.svelte';
	import PublicHeader from './PublicHeader.svelte';

	interface Props {
		children: Snippet;
		footerNote?: Snippet;
		pendingPath?: string | undefined;
		currentPath?: string;
	}

	let { children, footerNote, pendingPath, currentPath }: Props = $props();
</script>

<SkipLink />
<div class="flex min-h-dvh flex-col">
	<PublicHeader {currentPath} />
	<main id="konten" tabindex="-1" class="mx-auto w-full max-w-content flex-1 px-4 py-8 lg:px-8">
		<NavigationContent {pendingPath}>{@render children()}</NavigationContent>
	</main>
	<PublicFooter note={footerNote} />
</div>
