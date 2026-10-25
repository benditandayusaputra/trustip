<script lang="ts">
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import LayoutDashboard from '@lucide/svelte/icons/layout-dashboard';
	import Calculator from '@lucide/svelte/icons/calculator';
	import Package from '@lucide/svelte/icons/package';
	import Megaphone from '@lucide/svelte/icons/megaphone';
	import Search from '@lucide/svelte/icons/search';
	import { common as text } from '#lib/i18n/common.js';
	import { navigationFor, isCurrent, type OrganizationType } from './navigation';
	let {
		organizationType,
		currentPath
	}: { organizationType: OrganizationType; currentPath: string } = $props();
	const navigation = $derived(navigationFor(organizationType));
	const related = [
		{ href: '/app/hitung', label: text.nav.count, icon: Calculator },
		{ href: '/app/titipan', label: text.nav.drops, icon: Package },
		{ href: '/kebutuhan', label: text.nav.needsBoard, icon: Megaphone },
		{ href: '/cari/toko', label: text.nav.findStores, icon: Search },
		{ href: '/cari/penitip', label: text.nav.findConsignors, icon: Search }
	];
	const current = $derived(
		[...navigation.primary, ...navigation.more, ...related].find((item) =>
			isCurrent(item.href, currentPath)
		)
	);
	const Icon = $derived(current?.icon ?? LayoutDashboard);
	const searchHref = $derived(organizationType === 'STORE' ? '/cari/penitip' : '/cari/toko');
</script>

<nav aria-label={text.layout.breadcrumb} class="hidden min-w-0 items-center gap-4 lg:flex">
	<span
		class="flex size-11 items-center justify-center rounded-control border border-border text-text-muted"
	>
		<Icon class="size-5" aria-hidden="true" />
	</span>
	<div class="flex flex-col gap-1">
		<p class="text-xs text-text-muted">{text.layout.organizations[organizationType]}</p>
		<p class="flex items-center gap-2 text-sm font-semibold text-text">
			<a href="/app" class="hover:text-accent">{text.layout.navigationGroup}</a><ChevronRight
				class="size-4 text-text-muted"
				aria-hidden="true"
			/>{#if current && current.href !== currentPath}<a
					href={current.href}
					class="hover:text-accent">{current.label}</a
				>{:else}<span aria-current="page">{current?.label ?? text.layout.overview}</span>{/if}
		</p>
	</div>
</nav>
<a href={searchHref} class="workspace-search hidden xl:inline-flex">
	<Search class="size-4" aria-hidden="true" />{text.layout.searchPartners}
</a>
