<script lang="ts">
	import PublicShell from '#lib/components/layout/PublicShell.svelte';
	import { navigating, page } from '$app/state';
	import { pendingPagePath } from '#lib/components/layout/page-skeleton.js';
	import ApiVersion from '#lib/features/system/ApiVersion.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<PublicShell pendingPath={pendingPagePath(navigating.to?.url, page.url)}>
	{@render children()}
	{#snippet footerNote()}
		<ApiVersion />
	{/snippet}
</PublicShell>
