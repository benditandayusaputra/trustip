<script lang="ts">
	import PublicShell from '#lib/components/layout/PublicShell.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<PublicShell>
	{@render children()}
</PublicShell>
