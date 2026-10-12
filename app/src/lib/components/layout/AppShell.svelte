<script lang="ts">
	import type { Snippet } from 'svelte';
	import BottomNav from './BottomNav.svelte';
	import Sidebar from './Sidebar.svelte';
	import SkipLink from './SkipLink.svelte';
	import NavigationContent from './NavigationContent.svelte';
	import Brand from './Brand.svelte';
	import ThemePicker from '../ui/ThemePicker.svelte';
	import { common as text } from '#lib/i18n/common.js';
	import { navigationFor, type OrganizationType } from './navigation';

	interface Props {
		organizationType?: OrganizationType | undefined;
		currentPath: string;
		children: Snippet;
		actions?: Snippet;
		pendingPath?: string | undefined;
	}

	let { organizationType, currentPath, children, actions, pendingPath }: Props = $props();

	const navigation = $derived(organizationType ? navigationFor(organizationType) : undefined);
</script>

<SkipLink />
<div class="flex min-h-dvh">
	{#if organizationType && navigation}
		<Sidebar items={[...navigation.primary, ...navigation.more]} {currentPath} {organizationType} />
	{/if}
	<div class="flex min-w-0 flex-1 flex-col">
		<header class="border-b border-border bg-surface px-4 py-4 lg:px-8">
			<div class="mx-auto flex max-w-content flex-wrap items-center justify-between gap-2">
				{#if organizationType}
					<div class="lg:hidden"><Brand href="/app" /></div>
					<p class="hidden text-sm text-text-muted lg:block">{text.layout.workspace}</p>
				{:else}
					<Brand />
				{/if}
				<nav aria-label={text.layout.publicNavigation} class="flex flex-wrap items-center gap-2">
					{#if !organizationType}
						<a href="/cari/toko" class="hidden text-text-muted hover:text-text md:inline">
							{text.nav.findStores}
						</a>
						<a href="/kebutuhan" class="hidden text-text-muted hover:text-text md:inline">
							{text.nav.needsBoard}
						</a>
					{/if}
					<ThemePicker />{@render actions?.()}
				</nav>
			</div>
		</header>
		<main
			id="konten"
			tabindex="-1"
			class="mx-auto w-full max-w-content flex-1 px-4 pt-8 lg:px-8 lg:pb-8 {organizationType
				? 'pb-(--toast-offset)'
				: 'pb-8'}"
		>
			<NavigationContent {pendingPath}>{@render children()}</NavigationContent>
		</main>
		{#if !organizationType}
			<footer class="border-t border-border bg-surface px-4 py-6">
				<nav
					aria-label={text.layout.footerNavigation}
					class="mx-auto flex max-w-content flex-wrap gap-4 text-sm text-text-muted"
				>
					<a href="/privasi" class="hover:text-text">{text.nav.privacy}</a>
					<a href="/ketentuan" class="hover:text-text">{text.nav.terms}</a>
				</nav>
			</footer>
		{/if}
	</div>
</div>
{#if navigation}
	<BottomNav primary={navigation.primary} more={navigation.more} {currentPath} />
{/if}
