<script lang="ts">
	import type { Snippet } from 'svelte';
	import { afterNavigate } from '$app/navigation';
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
		isCurrent,
		isFocusPage,
		menuOfPage,
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
	const pageTitle = new PageTitle(() => organizationType !== undefined);
	setPageTitle(pageTitle);
	const items = $derived(navigation ? [...navigation.workspace, ...navigation.manage] : []);
	const inSection = (path: string) => items.some((item) => isCurrent(item.href, path));
	const pathOf = (href: string) => href.split('?')[0] ?? href;
	const menuOf = (path: string) =>
		organizationType ? menuOfPage(organizationType, path) : undefined;
	let origin = $state<string>();
	let previous = $state<string>();

	const hrefOf = (url: URL | null | undefined) => (url ? url.pathname + url.search : undefined);
	const isAnchor = (href: string | undefined) =>
		href !== undefined && (inSection(pathOf(href)) || isFocusPage(pathOf(href)));

	afterNavigate(({ from, to }) => {
		previous = hrefOf(from?.url);
		const target = hrefOf(to?.url);
		if (target === undefined || isAnchor(target)) {
			origin = undefined;
		} else if (isAnchor(previous)) {
			origin = previous;
		}
	});

	const listHref = $derived(pageTitle.entry?.backHref);

	function cameFromList(): boolean {
		return (
			listHref !== undefined && previous !== undefined && pathOf(previous) === pathOf(listHref)
		);
	}

	function parentOf(path: string): string | undefined {
		if (inSection(path)) {
			return cameFromList() ? previous : listHref;
		}
		return origin ?? listHref ?? menuOf(path) ?? items[0]?.href;
	}

	function menuPathOf(path: string, parent: string | undefined): string {
		if (inSection(path)) {
			return path;
		}
		const parentPath = parent === undefined ? path : pathOf(parent);
		if (inSection(parentPath)) {
			return parentPath;
		}
		return menuOf(parentPath) ?? menuOf(path) ?? path;
	}

	const backHref = $derived(parentOf(currentPath));
	const activePath = $derived(menuPathOf(currentPath, backHref));
</script>

<SkipLink />
<div class="flex min-h-dvh" data-shell={isPublic ? 'public' : 'workspace'}>
	{#if organizationType && navigation && !focused}
		<Sidebar {navigation} currentPath={activePath} {organizationType} />
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
						<WorkspaceBreadcrumb {organizationType} {currentPath} {backHref} {focused} />
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
			<NavigationContent {pendingPath} titled={isPublic}>{@render children()}</NavigationContent>
		</main>
		{#if isPublic}
			<PublicFooter />
		{/if}
	</div>
</div>
{#if bottomNavigation}
	<BottomNav
		primary={bottomNavigation.primary}
		more={bottomNavigation.more}
		currentPath={activePath}
	/>
{/if}
