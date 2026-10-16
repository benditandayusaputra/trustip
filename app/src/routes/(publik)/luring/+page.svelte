<script lang="ts">
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import WifiOff from '@lucide/svelte/icons/wifi-off';
	import { offline as id } from '#lib/i18n/offline.js';
</script>

<svelte:head>
	<title>{id.offline.title}</title>
</svelte:head>

<section class="mx-auto flex max-w-form flex-col items-center gap-4 py-8 text-center">
	<PageHeading icon={WifiOff}>{id.offline.title}</PageHeading>
	<p class="text-text-muted">{id.offline.body}</p>
	<a
		href="/"
		class="inline-flex min-h-12 items-center rounded-control bg-primary px-5 font-semibold text-on-accent"
	>
		{id.offline.retry}
	</a>
</section>
