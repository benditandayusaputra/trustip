<script lang="ts">
	import Menu from '@lucide/svelte/icons/menu';
	import Close from '@lucide/svelte/icons/x';
	import Store from '@lucide/svelte/icons/store';
	import PackageSearch from '@lucide/svelte/icons/package-search';
	import Megaphone from '@lucide/svelte/icons/megaphone';
	import BookOpen from '@lucide/svelte/icons/book-open';
	import CircleQuestionMark from '@lucide/svelte/icons/circle-question-mark';
	import ThemePicker from '../ui/ThemePicker.svelte';
	import { common as text } from '#lib/i18n/common.js';
	let menu: HTMLDetailsElement;
	let open = $state(false);
	function dismiss(event: MouseEvent) {
		if (open && event.target instanceof Node && !menu.contains(event.target)) open = false;
	}
	function closeWithEscape(event: KeyboardEvent) {
		if (event.key === 'Escape' && open) {
			open = false;
			menu.querySelector('summary')?.focus();
		}
	}
	const links = [
		{ href: '/cari/toko', label: text.nav.findStores, icon: Store },
		{ href: '/cari/penitip', label: text.nav.findConsignors, icon: PackageSearch },
		{ href: '/kebutuhan', label: text.nav.needsBoard, icon: Megaphone },
		{ href: '/panduan', label: text.nav.guide, icon: BookOpen },
		{ href: '/faq', label: text.nav.faq, icon: CircleQuestionMark }
	];
</script>

<svelte:window onclick={dismiss} onkeydown={closeWithEscape} />

<details bind:this={menu} bind:open class="public-menu md:hidden">
	<summary aria-label={text.layout.openMenu}>
		{#if open}<Close class="size-5" aria-hidden="true" />{:else}<Menu
				class="size-5"
				aria-hidden="true"
			/>{/if}
	</summary>
	<div class="public-menu-panel">
		<nav aria-label={text.layout.menu} class="flex flex-col gap-1">
			{#each links as link (link.href)}
				<a href={link.href} onclick={() => (open = false)}
					><link.icon class="size-5 shrink-0 text-accent" aria-hidden="true" />{link.label}</a
				>
			{/each}
		</nav>
		<div class="border-t border-border pt-4"><ThemePicker expanded /></div>
	</div>
</details>
