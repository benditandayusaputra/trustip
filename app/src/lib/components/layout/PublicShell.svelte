<script lang="ts">
	import type { Snippet } from 'svelte';
	import SkipLink from './SkipLink.svelte';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		children: Snippet;
		footerNote?: Snippet;
	}

	let { children, footerNote }: Props = $props();
</script>

<SkipLink />
<div class="flex min-h-dvh flex-col">
	<header class="border-b border-border bg-surface">
		<div class="mx-auto flex max-w-content items-center justify-between gap-4 px-4 py-2">
			<a href="/" class="text-lg font-semibold text-text">{id.brand}</a>
			<nav aria-label={id.layout.publicNavigation} class="flex items-center gap-4">
				<a href="/cari/toko" class="hidden text-text-muted hover:text-text md:inline">
					{id.nav.findStores}
				</a>
				<a href="/kebutuhan" class="hidden text-text-muted hover:text-text md:inline">
					{id.nav.needsBoard}
				</a>
				<a
					href="/masuk"
					class="inline-flex min-h-11 items-center rounded-control border border-border-strong px-4 font-semibold"
				>
					{id.nav.signIn}
				</a>
			</nav>
		</div>
	</header>
	<main id="konten" tabindex="-1" class="mx-auto w-full max-w-content flex-1 px-4 py-6">
		{@render children()}
	</main>
	<footer class="border-t border-border">
		<div
			class="mx-auto flex max-w-content flex-col gap-2 px-4 py-6 text-sm text-text-muted md:flex-row md:justify-between"
		>
			<nav aria-label={id.layout.footerNavigation} class="flex gap-4">
				<a href="/privasi" class="hover:text-text">{id.nav.privacy}</a>
				<a href="/ketentuan" class="hover:text-text">{id.nav.terms}</a>
			</nav>
			{@render footerNote?.()}
		</div>
	</footer>
</div>
