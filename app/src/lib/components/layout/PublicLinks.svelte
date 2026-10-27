<script lang="ts">
	import { common as text } from '#lib/i18n/common.js';
	import { isCurrent } from './navigation';

	let { currentPath = '' }: { currentPath?: string | undefined } = $props();

	const links = [
		{ href: '/', label: text.nav.home },
		{ href: '/cari/toko', label: text.nav.findStores },
		{ href: '/kebutuhan', label: text.nav.needsBoard },
		{ href: '/panduan', label: text.nav.guide }
	];
</script>

<div class="hidden items-center gap-1 md:flex">
	{#each links as link (link.href)}
		<a
			href={link.href}
			aria-current={(link.href === '/' ? currentPath === '/' : isCurrent(link.href, currentPath))
				? 'page'
				: undefined}
			class="public-nav-link">{link.label}</a
		>
	{/each}
</div>
