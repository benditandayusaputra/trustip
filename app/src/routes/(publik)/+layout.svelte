<script lang="ts">
	import PublicShell from '#lib/components/layout/PublicShell.svelte';
	import { navigating, page } from '$app/state';
	import { pendingPagePath } from '#lib/components/layout/page-skeleton.js';
	import ApiVersion from '#lib/features/system/ApiVersion.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();

	const backTargets = [
		['/t/', '/cari/toko'],
		['/u/', '/cari/penitip'],
		['/p/', '/cari/penitip'],
		['/kebutuhan/', '/kebutuhan']
	];
	const backHref = $derived(
		backTargets.find(([prefix]) => page.url.pathname.startsWith(prefix ?? ''))?.[1]
	);
</script>

<PublicShell
	pendingPath={pendingPagePath(navigating.to?.url, page.url)}
	currentPath={page.url.pathname}
	{backHref}
>
	{@render children()}
	{#snippet footerNote()}
		<ApiVersion />
	{/snippet}
</PublicShell>
