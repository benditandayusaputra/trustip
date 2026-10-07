<script lang="ts">
	import Ellipsis from '@lucide/svelte/icons/ellipsis';
	import BottomSheet from '../ui/BottomSheet.svelte';
	import NavLink from './NavLink.svelte';
	import type { NavItem } from './navigation';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		primary: NavItem[];
		more: NavItem[];
		currentPath: string;
	}

	let { primary, more, currentPath }: Props = $props();

	let moreOpen = $state(false);
</script>

<nav
	aria-label={id.layout.mainNavigation}
	class="fixed inset-x-0 bottom-0 z-(--z-nav) flex h-(--bottom-nav-height) border-t border-border bg-surface lg:hidden"
>
	{#each primary as item (item.href)}
		<NavLink {item} {currentPath} layout="tab" />
	{/each}
	<button
		type="button"
		aria-haspopup="dialog"
		onclick={() => (moreOpen = true)}
		class="flex min-h-11 flex-1 flex-col items-center justify-center gap-1 text-xs text-text-muted"
	>
		<Ellipsis class="size-6" aria-hidden="true" />
		<span>{id.layout.more}</span>
	</button>
</nav>

<BottomSheet bind:open={moreOpen} title={id.layout.more}>
	<ul class="flex flex-col">
		{#each more as item (item.href)}
			<li><NavLink {item} {currentPath} layout="row" /></li>
		{/each}
	</ul>
</BottomSheet>
