<script lang="ts">
	import PublicShell from '#lib/components/layout/PublicShell.svelte';
	import ApiVersion from '#lib/features/system/ApiVersion.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<PublicShell>
	{@render children()}
	{#snippet footerNote()}
		<ApiVersion />
	{/snippet}
</PublicShell>
