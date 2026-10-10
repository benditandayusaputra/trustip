<script lang="ts">
	import NavLink from './NavLink.svelte';
	import Brand from './Brand.svelte';
	import type { NavItem, OrganizationType } from './navigation';
	import { common as id } from '#lib/i18n/common.js';

	interface Props {
		items: NavItem[];
		currentPath: string;
		organizationType: OrganizationType;
	}

	let { items, currentPath, organizationType }: Props = $props();
</script>

<aside
	class="sticky top-0 hidden h-dvh w-(--sidebar-width) shrink-0 flex-col gap-6 overflow-y-auto border-r border-border bg-surface p-4 lg:flex"
>
	<Brand />
	<div class="rounded-control bg-surface-hover p-4">
		<p class="text-sm font-semibold text-text">{id.layout.organizations[organizationType]}</p>
		<p class="text-xs text-text-muted">{id.layout.sharedRecords}</p>
	</div>
	<nav aria-label={id.layout.mainNavigation}>
		<ul class="flex flex-col gap-1">
			{#each items as item (item.href)}
				<li><NavLink {item} {currentPath} layout="row" /></li>
			{/each}
		</ul>
	</nav>
	<p class="mt-auto px-4 text-xs text-text-muted">{id.layout.workspace}</p>
</aside>
