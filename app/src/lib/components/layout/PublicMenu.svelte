<script lang="ts">
	import Menu from '@lucide/svelte/icons/menu';
	import Store from '@lucide/svelte/icons/store';
	import PackageSearch from '@lucide/svelte/icons/package-search';
	import Megaphone from '@lucide/svelte/icons/megaphone';
	import { common as text } from '#lib/i18n/common.js';
	let menu: HTMLDetailsElement;
	function dismiss(event: MouseEvent) {
		if (menu.open && event.target instanceof Node && !menu.contains(event.target))
			menu.open = false;
	}
	function closeWithEscape(event: KeyboardEvent) {
		if (event.key === 'Escape' && menu.open) {
			menu.open = false;
			menu.querySelector('summary')?.focus();
		}
	}
	const links = [
		{ href: '/cari/toko', label: text.nav.findStores, icon: Store },
		{ href: '/cari/penitip', label: text.nav.findConsignors, icon: PackageSearch },
		{ href: '/kebutuhan', label: text.nav.needsBoard, icon: Megaphone }
	];
</script>

<svelte:window onclick={dismiss} onkeydown={closeWithEscape} />

<details bind:this={menu} class="public-menu md:hidden">
	<summary aria-label={text.layout.openMenu}><Menu class="size-5" aria-hidden="true" /></summary>
	<nav aria-label={text.layout.menu} class="public-menu-panel">
		{#each links as link (link.href)}
			<a href={link.href} onclick={() => (menu.open = false)}
				><link.icon class="size-5 shrink-0 text-accent" aria-hidden="true" />{link.label}</a
			>
		{/each}
	</nav>
</details>
