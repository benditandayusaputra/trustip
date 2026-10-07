<script lang="ts">
	import WifiOff from '@lucide/svelte/icons/wifi-off';
	import { id } from '#lib/i18n/id.js';
</script>

<svelte:head>
	<title>{id.offline.title}</title>
</svelte:head>

<section class="mx-auto flex max-w-form flex-col items-center gap-4 py-8 text-center">
	<WifiOff class="size-12 text-warning" aria-hidden="true" />
	<h1 class="text-xl">{id.offline.title}</h1>
	<p class="text-text-muted">{id.offline.body}</p>
	<a
		href="/"
		class="inline-flex min-h-12 items-center rounded-control bg-accent px-5 font-semibold text-on-accent"
	>
		{id.offline.retry}
	</a>
</section>
