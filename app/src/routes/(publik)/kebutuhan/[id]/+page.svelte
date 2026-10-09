<script lang="ts">
	import { page } from '$app/state';
	import LinkPreview from '#lib/components/ui/LinkPreview.svelte';
	import PublicNeedPage from '#lib/features/directory/components/PublicNeedPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<LinkPreview
	title={data.need.title}
	description={data.need.description}
	url={page.url.href}
	image={null}
/>
<PublicNeedPage need={data.need} shareUrl={page.url.href} />
