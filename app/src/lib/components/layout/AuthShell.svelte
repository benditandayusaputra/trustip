<script lang="ts">
	import type { Snippet } from 'svelte';
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import Brand from './Brand.svelte';
	import SkipLink from './SkipLink.svelte';
	import NavigationContent from './NavigationContent.svelte';
	import TradeIllustration from '../ui/TradeIllustration.svelte';
	import ThemePicker from '../ui/ThemePicker.svelte';
	import { common as id } from '#lib/i18n/common.js';

	interface Props {
		children: Snippet;
		pendingPath?: string | undefined;
	}
	let { children, pendingPath }: Props = $props();
	const text = id.auth.presentation;
</script>

<SkipLink />
<div class="flex min-h-dvh flex-col">
	<header
		class="mx-auto flex w-full max-w-content items-center justify-between gap-4 px-4 py-4 lg:px-8"
	>
		<Brand />
		<div class="flex items-center gap-4">
			<a
				href="/"
				class="hidden min-h-11 items-center gap-2 text-sm font-medium text-text-muted hover:text-text md:inline-flex"
			>
				<ArrowLeft class="size-4" aria-hidden="true" />{text.backHome}
			</a>
			<ThemePicker />
		</div>
	</header>
	<main
		id="konten"
		tabindex="-1"
		class="mx-auto grid w-full max-w-content min-w-0 flex-1 grid-cols-1 items-center gap-6 px-4 py-4 lg:grid-cols-2 lg:gap-12 lg:px-8 lg:py-8"
	>
		<aside
			class="auth-presentation flex min-w-0 flex-col gap-4 rounded-card p-4 lg:gap-6 lg:p-8"
			aria-label={text.title}
		>
			<div class="hidden flex-col gap-4 lg:flex">
				<p class="section-label">{text.eyebrow}</p>
				<h2 class="text-hero-compact font-semibold">
					{text.headline}<br /><span class="auth-accent">{text.headlineAccent}</span>
				</h2>
				<p class="auth-copy text-base">{text.body}</p>
			</div>
			<TradeIllustration compact />
			<p class="auth-copy text-center text-sm font-medium lg:text-left">{text.caption}</p>
		</aside>
		<div class="auth-form mx-auto w-full max-w-auth-form min-w-0">
			<NavigationContent {pendingPath}>{@render children()}</NavigationContent>
		</div>
	</main>
	<footer
		class="mx-auto flex w-full max-w-content flex-wrap items-center justify-between gap-4 px-4 py-6 text-xs text-text-muted lg:px-8"
	>
		<p>{text.footer}</p>
		<nav aria-label={id.layout.footerNavigation} class="flex flex-wrap gap-4">
			<a href="/privasi" class="inline-flex min-h-11 items-center hover:text-text"
				>{id.nav.privacy}</a
			>
			<a href="/ketentuan" class="inline-flex min-h-11 items-center hover:text-text"
				>{id.nav.terms}</a
			>
		</nav>
	</footer>
</div>
