<script lang="ts">
	import PublicStoreProfile from '#lib/features/directory/components/PublicStoreProfile.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<svelte:head>
	<title>{data.store.name}</title>
	<meta name="description" content={data.store.description} />
</svelte:head>

<PublicStoreProfile store={data.store} />
