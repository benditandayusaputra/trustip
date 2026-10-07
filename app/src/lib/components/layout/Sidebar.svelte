<script lang="ts">
	import NavLink from './NavLink.svelte';
	import type { NavItem } from './navigation';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		items: NavItem[];
		currentPath: string;
	}

	let { items, currentPath }: Props = $props();
</script>

<aside
	class="sticky top-0 hidden h-dvh w-(--sidebar-width) shrink-0 flex-col gap-6 border-r border-border bg-surface p-4 lg:flex"
>
	<a href="/" class="px-4 text-lg font-semibold text-text">{id.brand}</a>
	<nav aria-label={id.layout.mainNavigation}>
		<ul class="flex flex-col gap-1">
			{#each items as item (item.href)}
				<li><NavLink {item} {currentPath} layout="row" /></li>
			{/each}
		</ul>
	</nav>
</aside>
