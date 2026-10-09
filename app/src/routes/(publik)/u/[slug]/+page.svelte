<script lang="ts">
	import { page } from '$app/state';
	import LinkPreview from '#lib/components/ui/LinkPreview.svelte';
	import PublicConsignorProfile from '#lib/features/directory/components/PublicConsignorProfile.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<LinkPreview
	title={data.consignor.name}
	description={data.consignor.description}
	url={page.url.href}
	image={data.consignor.logo_url}
/>

<PublicConsignorProfile shareUrl={page.url.href} consignor={data.consignor} />
