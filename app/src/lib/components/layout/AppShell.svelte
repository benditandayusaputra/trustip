<script lang="ts">
	import type { Snippet } from 'svelte';
	import BottomNav from './BottomNav.svelte';
	import Sidebar from './Sidebar.svelte';
	import SkipLink from './SkipLink.svelte';
	import NavigationContent from './NavigationContent.svelte';
	import PublicMenu from './PublicMenu.svelte';
	import Brand from './Brand.svelte';
	import WorkspaceBreadcrumb from './WorkspaceBreadcrumb.svelte';
	import PublicLinks from './PublicLinks.svelte';
	import ThemePicker from '../ui/ThemePicker.svelte';
	import { common as text } from '#lib/i18n/common.js';
	import { bottomNavigationOf, navigationFor, type OrganizationType } from './navigation';
	import { PageTitle, setPageTitle } from './page-title.svelte';

	interface Props {
		organizationType?: OrganizationType | undefined;
		currentPath: string;
		children: Snippet;
		actions?: Snippet;
		pendingPath?: string | undefined;
		workspace?: boolean;
	}

	let {
		organizationType,
		currentPath,
		children,
		actions,
		pendingPath,
		workspace = false
	}: Props = $props();

	const navigation = $derived(organizationType ? navigationFor(organizationType) : undefined);
	const isPublic = $derived(!organizationType && !workspace);
	const bottomNavigation = $derived(navigation ? bottomNavigationOf(navigation) : undefined);
	setPageTitle(new PageTitle(() => organizationType !== undefined));
</script>

<SkipLink />
<div class="flex min-h-dvh">
	{#if organizationType && navigation}
		<Sidebar {navigation} {currentPath} {organizationType} />
	{:else if workspace}
		<div
			class="workspace-sidebar sticky top-0 hidden h-dvh w-(--sidebar-width) shrink-0 p-5 lg:block"
		>
			<div class="px-2 py-2"><Brand href="/app" /></div>
		</div>
	{/if}
	<div class="flex min-w-0 flex-1 flex-col">
		<header class="workspace-header flex items-center px-4 py-2 lg:px-8">
			<div class="mx-auto flex w-full max-w-content flex-wrap items-center justify-between gap-4">
				{#if organizationType}
					<WorkspaceBreadcrumb {organizationType} {currentPath} />
				{:else if workspace}
					<div class="lg:hidden"><Brand href="/app" /></div>
				{:else}
					<Brand compact />
				{/if}
				<nav aria-label={text.layout.publicNavigation} class="flex flex-wrap items-center gap-2">
					{#if isPublic}<PublicLinks {currentPath} /><PublicMenu />{/if}<ThemePicker
					/>{@render actions?.()}
				</nav>
			</div>
		</header>
		<main
			id="konten"
			tabindex="-1"
			class="workspace-main mx-auto w-full max-w-content flex-1 px-4 pt-8 lg:px-8 lg:pb-8 {isPublic
				? 'pb-8'
				: 'pb-(--toast-offset)'}"
		>
			<NavigationContent {pendingPath}>{@render children()}</NavigationContent>
		</main>
		{#if isPublic}
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
{#if bottomNavigation}
	<BottomNav primary={bottomNavigation.primary} more={bottomNavigation.more} {currentPath} />
{/if}
