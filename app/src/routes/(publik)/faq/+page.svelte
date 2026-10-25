<script lang="ts">
	import FaqPage from '#lib/features/help/components/FaqPage.svelte';
	import { faq } from '#lib/i18n/faq.js';
</script>

<svelte:head>
	<title>{faq.title}</title>
	<meta name="description" content={faq.intro} />
</svelte:head>

<FaqPage />
