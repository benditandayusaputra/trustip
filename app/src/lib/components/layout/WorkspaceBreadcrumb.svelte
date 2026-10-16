<script lang="ts">
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import LayoutDashboard from '@lucide/svelte/icons/layout-dashboard';
	import Search from '@lucide/svelte/icons/search';
	import { common as text } from '#lib/i18n/common.js';
	import { navigationFor, isCurrent, type OrganizationType } from './navigation';
	let {
		organizationType,
		currentPath
	}: { organizationType: OrganizationType; currentPath: string } = $props();
	const navigation = $derived(navigationFor(organizationType));
	const current = $derived(
		[...navigation.primary, ...navigation.more].find((item) => isCurrent(item.href, currentPath))
	);
	const Icon = $derived(current?.icon ?? LayoutDashboard);
	const searchHref = $derived(organizationType === 'STORE' ? '/cari/penitip' : '/cari/toko');
</script>

<div class="hidden min-w-0 items-center gap-4 lg:flex">
	<span
		class="flex size-11 items-center justify-center rounded-control border border-border text-text-muted"
	>
		<Icon class="size-5" aria-hidden="true" />
	</span>
	<div class="flex flex-col gap-1">
		<p class="text-xs text-text-muted">{text.layout.organizations[organizationType]}</p>
		<p class="flex items-center gap-2 text-sm font-semibold text-text">
			{text.layout.navigationGroup}<ChevronRight
				class="size-4 text-text-muted"
				aria-hidden="true"
			/>{current?.label ?? text.layout.overview}
		</p>
	</div>
</div>
<a href={searchHref} class="workspace-search hidden xl:inline-flex">
	<Search class="size-4" aria-hidden="true" />{text.layout.searchPartners}
</a>
