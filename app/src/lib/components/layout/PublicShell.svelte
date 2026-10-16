<script lang="ts">
	import type { Snippet } from 'svelte';
	import ArrowUpRight from '@lucide/svelte/icons/arrow-up-right';
	import Store from '@lucide/svelte/icons/store';
	import Megaphone from '@lucide/svelte/icons/megaphone';
	import SkipLink from './SkipLink.svelte';
	import NavigationContent from './NavigationContent.svelte';
	import PublicMenu from './PublicMenu.svelte';
	import Brand from './Brand.svelte';
	import PublicFooter from './PublicFooter.svelte';
	import ThemePicker from '../ui/ThemePicker.svelte';
	import { common as id } from '#lib/i18n/common.js';

	interface Props {
		children: Snippet;
		footerNote?: Snippet;
		actions?: Snippet;
		pendingPath?: string | undefined;
	}

	let { children, footerNote, actions, pendingPath }: Props = $props();
</script>

<SkipLink />
<div class="flex min-h-dvh flex-col">
	<header class="public-header">
		<div
			class="mx-auto flex max-w-content flex-wrap items-center justify-between gap-2 px-4 py-4 lg:px-8"
		>
			<Brand compact />
			<nav
				aria-label={id.layout.publicNavigation}
				class="flex flex-wrap items-center gap-2 md:gap-4"
			>
				<a href="/cari/toko" class="public-nav-link hidden md:inline-flex">
					<Store class="size-4" aria-hidden="true" />{id.nav.findStores}
				</a>
				<a href="/kebutuhan" class="public-nav-link hidden md:inline-flex">
					<Megaphone class="size-4" aria-hidden="true" />{id.nav.needsBoard}
				</a>
				<PublicMenu /><ThemePicker />
				{#if actions}
					{@render actions()}
				{:else}
					<a
						href="/masuk"
						class="inline-flex min-h-11 items-center gap-2 rounded-control bg-primary px-4 font-semibold text-on-accent hover:bg-primary-hover"
					>
						{id.nav.signIn}<ArrowUpRight class="size-4" aria-hidden="true" />
					</a>
				{/if}
			</nav>
		</div>
	</header>
	<main id="konten" tabindex="-1" class="mx-auto w-full max-w-content flex-1 px-4 py-8 lg:px-8">
		<NavigationContent {pendingPath}>{@render children()}</NavigationContent>
	</main>
	<PublicFooter note={footerNote} />
</div>
