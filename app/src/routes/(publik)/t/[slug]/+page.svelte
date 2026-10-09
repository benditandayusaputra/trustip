<script lang="ts">
	import { page } from '$app/state';
	import LinkPreview from '#lib/components/ui/LinkPreview.svelte';
	import PublicStoreProfile from '#lib/features/directory/components/PublicStoreProfile.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<LinkPreview
	title={data.store.name}
	description={data.store.description}
	url={page.url.href}
	image={data.store.logo_url}
/>

<PublicStoreProfile shareUrl={page.url.href} store={data.store} />
