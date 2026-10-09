<script lang="ts">
	import PublicConsignorProfile from '#lib/features/directory/components/PublicConsignorProfile.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<svelte:head>
	<title>{data.consignor.name}</title>
	<meta name="description" content={data.consignor.description} />
</svelte:head>

<PublicConsignorProfile consignor={data.consignor} />
