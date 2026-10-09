<script lang="ts">
	import PublicProductPage from '#lib/features/directory/components/PublicProductPage.svelte';
	import { locateDevice } from '#lib/features/organization/geolocation.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<svelte:head>
	<title>{data.product.name}</title>
	<meta name="description" content={data.product.description} />
</svelte:head>

<PublicProductPage
	product={data.product}
	consignorSlug={data.consignorSlug}
	locate={locateDevice}
/>
