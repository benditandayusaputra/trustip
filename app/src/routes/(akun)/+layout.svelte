<script lang="ts">
	import AuthShell from '#lib/components/layout/AuthShell.svelte';
	import { navigating, page } from '$app/state';
	import { pendingPagePath } from '#lib/components/layout/page-skeleton.js';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<AuthShell pendingPath={pendingPagePath(navigating.to?.url, page.url)}>
	{@render children()}
</AuthShell>
