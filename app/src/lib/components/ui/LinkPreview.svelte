<script lang="ts">
	interface Props {
		title: string;
		description: string;
		url: string;
		image: string | null;
	}

	let { title, description, url, image }: Props = $props();

	const fallbackImage = '/icons/icon-512.png';
	const imageUrl = $derived(image ?? new URL(fallbackImage, url).toString());
</script>

<svelte:head>
	<title>{title}</title>
	<meta name="description" content={description} />
	<meta property="og:type" content="website" />
	<meta property="og:title" content={title} />
	<meta property="og:description" content={description} />
	<meta property="og:url" content={url} />
	<meta property="og:image" content={imageUrl} />
	<meta name="twitter:card" content="summary_large_image" />
</svelte:head>
