<script lang="ts">
	import type { Snippet } from 'svelte';
	import ChevronLeft from '@lucide/svelte/icons/chevron-left';
	import Brand from './Brand.svelte';
	import SkipLink from './SkipLink.svelte';
	import NavigationContent from './NavigationContent.svelte';
	import TradeIllustration from '../ui/TradeIllustration.svelte';
	import { common as id } from '#lib/i18n/common.js';

	interface Props {
		children: Snippet;
		pendingPath?: string | undefined;
	}
	let { children, pendingPath }: Props = $props();
	const text = id.auth.presentation;
</script>

<SkipLink />
<div class="grid min-h-dvh grid-cols-1 lg:grid-cols-2">
	<aside
		class="auth-presentation flex min-w-0 flex-col gap-4 px-4 py-2 lg:sticky lg:top-0 lg:h-dvh lg:gap-6 lg:px-5 lg:py-4"
		aria-label={text.title}
	>
		<header class="flex items-center gap-1">
			<a
				href="/"
				aria-label={text.backHome}
				class="flex size-11 shrink-0 items-center justify-center rounded-control hover:bg-sidebar-hover"
			>
				<ChevronLeft class="size-8" aria-hidden="true" />
			</a>
			<Brand />
		</header>
		<div
			class="hidden flex-1 flex-col items-center justify-center gap-6 text-center lg:flex lg:pb-12"
		>
			<div class="flex flex-col items-center gap-6">
				<p class="auth-eyebrow rounded-pill px-4 py-1 text-xs font-semibold uppercase">
					{text.eyebrow}
				</p>
				<h2 class="text-hero-compact font-bold">
					{text.headline}<br />{text.headlineAccent}
				</h2>
				<p class="auth-copy max-w-reading text-base">{text.body}</p>
			</div>
			<div class="w-full"><TradeIllustration compact /></div>
			<p class="auth-copy text-sm font-medium">{text.caption}</p>
		</div>
	</aside>
	<main
		id="konten"
		tabindex="-1"
		class="flex min-w-0 items-center justify-center px-4 py-6 lg:px-8 lg:py-8"
	>
		<div class="w-full min-w-0">
			<NavigationContent {pendingPath}>{@render children()}</NavigationContent>
		</div>
	</main>
</div>
