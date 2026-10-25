<script lang="ts">
	import GuidePage from '#lib/features/help/components/GuidePage.svelte';
	import { guide } from '#lib/i18n/guide.js';
</script>

<svelte:head>
	<title>{guide.title}</title>
	<meta name="description" content={guide.intro} />
</svelte:head>

<GuidePage />
