<script lang="ts">
	import ChevronLeft from '@lucide/svelte/icons/chevron-left';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import LayoutDashboard from '@lucide/svelte/icons/layout-dashboard';
	import { common as text } from '#lib/i18n/common.js';
	import {
		navigationFor,
		isCurrent,
		sectionsOutsideNavigation,
		type OrganizationType
	} from './navigation';
	import { getPageTitle } from './page-title.svelte';
	let {
		organizationType,
		currentPath,
		focused = false
	}: { organizationType: OrganizationType; currentPath: string; focused?: boolean } = $props();
	const pageTitle = getPageTitle();
	const navigation = $derived(navigationFor(organizationType));
	const items = $derived([...navigation.workspace, ...navigation.manage]);
	const section = $derived(items.find((item) => isCurrent(item.href, currentPath)));
	const entry = $derived(pageTitle?.entry);
	const backHref = $derived(entry?.backHref ?? (section ? undefined : items[0]?.href));
	const parent = $derived(
		backHref === undefined
			? undefined
			: [...items, ...sectionsOutsideNavigation].find((item) => isCurrent(item.href, backHref))
	);
	const Icon = $derived(entry?.icon ?? section?.icon ?? LayoutDashboard);
</script>

<div class="flex min-w-0 items-center gap-2">
	{#if backHref}
		<a
			href={backHref}
			aria-label={text.layout.back}
			class="flex size-11 shrink-0 items-center justify-center rounded-control border border-border text-text hover:border-accent hover:text-accent"
			><ChevronLeft class="size-5" aria-hidden="true" /></a
		>
	{:else}
		<span
			class="flex size-11 shrink-0 items-center justify-center rounded-control border border-border text-text-muted"
			><Icon class="size-5" aria-hidden="true" /></span
		>
	{/if}
	<div class="flex min-w-0 flex-col">
		{#if !focused}
			<p class="hidden text-xs text-text-muted compact:block">
				{text.layout.organizations[organizationType]}
			</p>
		{/if}
		<div class="flex min-w-0 items-center gap-1 text-sm font-semibold">
			{#if backHref && parent && !focused}
				<a href={backHref} class="hidden shrink-0 text-text hover:text-accent compact:inline"
					>{parent.label}</a
				><ChevronRight
					class="hidden size-4 shrink-0 text-text-muted compact:block"
					aria-hidden="true"
				/>
			{/if}
			{#if entry}
				<h1 class="truncate text-text">{@render entry.title()}</h1>
			{:else}
				<span aria-current="page" class="truncate text-text"
					>{section?.label ?? text.layout.overview}</span
				>
			{/if}
		</div>
	</div>
</div>
