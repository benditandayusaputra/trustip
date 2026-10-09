<script lang="ts">
	import { page } from '$app/state';
	import LinkPreview from '#lib/components/ui/LinkPreview.svelte';
	import PublicProductPage from '#lib/features/directory/components/PublicProductPage.svelte';
	import { locateDevice } from '#lib/features/organization/geolocation.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<LinkPreview
	title={data.product.name}
	description={data.product.description}
	url={page.url.href}
	image={data.product.photo_url}
/>

<PublicProductPage
	product={data.product}
	consignorSlug={data.consignorSlug}
	locate={locateDevice}
	shareUrl={page.url.href}
/>
