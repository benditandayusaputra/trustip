<script lang="ts">
	import NavLink from './NavLink.svelte';
	import Brand from './Brand.svelte';
	import type { Navigation, OrganizationType } from './navigation';
	import { common as id } from '#lib/i18n/common.js';
	interface Props {
		navigation: Navigation;
		currentPath: string;
		organizationType: OrganizationType;
	}
	let { navigation, currentPath, organizationType }: Props = $props();
	const groups = $derived([
		{ label: id.layout.navigationGroup, items: navigation.workspace },
		{ label: id.layout.managementGroup, items: navigation.manage }
	]);
</script>

<aside
	class="workspace-sidebar sticky top-0 hidden h-dvh w-(--sidebar-width) shrink-0 flex-col gap-6 overflow-y-auto p-5 lg:flex"
>
	<div class="px-2 py-2">
		<Brand href="/app" subtitle={id.layout.organizations[organizationType]} />
	</div>
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
</aside>
