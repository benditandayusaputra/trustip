<script lang="ts">
	import type { Snippet } from 'svelte';
	import BottomNav from './BottomNav.svelte';
	import Sidebar from './Sidebar.svelte';
	import SkipLink from './SkipLink.svelte';
	import NavigationContent from './NavigationContent.svelte';
	import Brand from './Brand.svelte';
	import WorkspaceBreadcrumb from './WorkspaceBreadcrumb.svelte';
	import PublicFooter from './PublicFooter.svelte';
	import PublicHeader from './PublicHeader.svelte';
	import ThemePicker from '../ui/ThemePicker.svelte';
	import { common as text } from '#lib/i18n/common.js';
	import {
		bottomNavigationOf,
		isFocusPage,
		navigationFor,
		type OrganizationType
	} from './navigation';
	import { PageTitle, setPageTitle } from './page-title.svelte';

	interface Props {
		organizationType?: OrganizationType | undefined;
		currentPath: string;
		children: Snippet;
		actions?: Snippet | undefined;
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
	const focused = $derived(isFocusPage(currentPath));
	const bottomNavigation = $derived(navigation ? bottomNavigationOf(navigation) : undefined);
	setPageTitle(new PageTitle(() => organizationType !== undefined));
</script>

<SkipLink />
<div class="flex min-h-dvh">
	{#if organizationType && navigation && !focused}
		<Sidebar {navigation} {currentPath} {organizationType} />
	{:else if workspace && !organizationType}
		<div
			class="workspace-sidebar sticky top-0 hidden h-dvh w-(--sidebar-width) shrink-0 p-5 lg:block"
		>
			<div class="px-2 py-2"><Brand href="/app" /></div>
		</div>
	{/if}
	<div class="flex min-w-0 flex-1 flex-col">
		{#if isPublic}
			<PublicHeader {currentPath} {actions} />
		{:else}
			<header class="workspace-header flex items-center px-4 py-2 lg:px-8">
				<div class="mx-auto flex w-full max-w-content flex-wrap items-center justify-between gap-4">
					{#if organizationType}
						<WorkspaceBreadcrumb {organizationType} {currentPath} {focused} />
					{:else}
						<div class="lg:hidden"><Brand href="/app" /></div>
					{/if}
					<nav aria-label={text.layout.publicNavigation} class="flex flex-wrap items-center gap-2">
						<ThemePicker />{@render actions?.()}
					</nav>
				</div>
			</header>
		{/if}
		<main
			id="konten"
			tabindex="-1"
			class="workspace-main isolate mx-auto w-full max-w-content flex-1 px-4 pt-8 lg:px-8 lg:pb-8 {isPublic
				? 'pb-8'
				: 'pb-(--toast-offset)'}"
		>
			<NavigationContent {pendingPath}>{@render children()}</NavigationContent>
		</main>
		{#if isPublic}
			<PublicFooter />
		{/if}
	</div>
</div>
{#if bottomNavigation}
	<BottomNav primary={bottomNavigation.primary} more={bottomNavigation.more} {currentPath} />
{/if}
