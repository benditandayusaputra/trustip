<script lang="ts">
	import type { Snippet } from 'svelte';
	import ArrowUpRight from '@lucide/svelte/icons/arrow-up-right';
	import ChevronLeft from '@lucide/svelte/icons/chevron-left';
	import SkipLink from './SkipLink.svelte';
	import NavigationContent from './NavigationContent.svelte';
	import PublicMenu from './PublicMenu.svelte';
	import Brand from './Brand.svelte';
	import PublicFooter from './PublicFooter.svelte';
	import PublicLinks from './PublicLinks.svelte';
	import ThemePicker from '../ui/ThemePicker.svelte';
	import { common as id } from '#lib/i18n/common.js';

	interface Props {
		children: Snippet;
		footerNote?: Snippet;
		actions?: Snippet;
		pendingPath?: string | undefined;
		currentPath?: string;
		backHref?: string | undefined;
	}

	let { children, footerNote, actions, pendingPath, currentPath, backHref }: Props = $props();

	function goBack(event: MouseEvent) {
		if (history.length > 1) {
			event.preventDefault();
			history.back();
		}
	}
</script>

<SkipLink />
<div class="flex min-h-dvh flex-col">
	<header class="public-header">
		<div class="mx-auto flex max-w-content items-center justify-between gap-4 px-4 py-2 lg:px-8">
			<Brand />
			<nav aria-label={id.layout.publicNavigation} class="flex items-center gap-2">
				<PublicLinks {currentPath} />
			</nav>
			<div class="flex items-center gap-2">
				<div class="hidden md:block"><ThemePicker /></div>
				{#if actions}
					{@render actions()}
				{:else}
					<a
						href="/masuk"
						class="inline-flex min-h-11 items-center gap-2 rounded-control bg-primary px-4 text-sm font-semibold text-on-accent hover:bg-primary-hover"
					>
						{id.nav.signIn}<ArrowUpRight class="size-4" aria-hidden="true" />
					</a>
				{/if}
				<PublicMenu />
			</div>
		</div>
	</header>
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
