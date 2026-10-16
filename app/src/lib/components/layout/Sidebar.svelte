<script lang="ts">
	import ShieldCheck from '@lucide/svelte/icons/shield-check';
	import Store from '@lucide/svelte/icons/store';
	import ChefHat from '@lucide/svelte/icons/chef-hat';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
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
	const primaryCount = 4;
	const groups = $derived([
		{ label: id.layout.navigationGroup, items: items.slice(0, primaryCount) },
		{ label: id.layout.managementGroup, items: items.slice(primaryCount) }
	]);
	const OrganizationIcon = $derived(organizationType === 'STORE' ? Store : ChefHat);
</script>

<aside
	class="workspace-sidebar sticky top-0 hidden h-dvh w-(--sidebar-width) shrink-0 flex-col gap-6 overflow-y-auto p-5 lg:flex"
>
	<div class="px-2 py-2"><Brand href="/app" /></div>
	<a
		href="/app/pengaturan"
		class="flex items-center gap-2 rounded-control border border-sidebar-border bg-sidebar-hover p-4"
	>
		<OrganizationIcon class="size-5 shrink-0 text-sidebar-accent" aria-hidden="true" />
		<span class="min-w-0 flex-1 text-sm font-semibold"
			>{id.layout.organizations[organizationType]}</span
		>
		<ChevronRight class="size-4 shrink-0 text-sidebar-muted" aria-hidden="true" />
	</a>
	<nav aria-label={id.layout.mainNavigation} class="flex flex-col gap-6">
		{#each groups as group (group.label)}
			<div class="flex flex-col gap-2">
				<p
					class="px-4 text-xs font-medium tracking-(--letter-spacing-label) text-sidebar-muted uppercase"
				>
					{group.label}
				</p>
				<ul class="flex flex-col gap-1">
					{#each group.items as item (item.href)}
						<li><NavLink {item} {currentPath} layout="row" /></li>
					{/each}
				</ul>
			</div>
		{/each}
	</nav>
	<div
		class="mt-auto flex flex-col gap-2 rounded-card border border-sidebar-border bg-sidebar-hover p-4"
	>
		<ShieldCheck class="mb-2 size-6 text-sidebar-accent" aria-hidden="true" />
		<p class="text-sm font-semibold">{id.layout.secureTitle}</p>
		<p class="text-xs leading-body text-sidebar-muted">{id.layout.secureBody}</p>
	</div>
</aside>
