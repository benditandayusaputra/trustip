<script lang="ts">
	import { page } from '$app/state';
	import { id } from '#lib/i18n/id.js';

	const notFoundStatus = 404;
	const isNotFound = $derived(page.status === notFoundStatus);
</script>

<svelte:head>
	<title>{isNotFound ? id.errorPage.notFoundTitle : id.errorPage.failedTitle}</title>
</svelte:head>

<main id="konten" class="mx-auto flex min-h-dvh max-w-form flex-col justify-center gap-4 px-4">
	<p class="font-mono text-kpi font-bold text-accent">{page.status}</p>
	<h1 class="text-xl">{isNotFound ? id.errorPage.notFoundTitle : id.errorPage.failedTitle}</h1>
	<p class="text-text-muted">
		{isNotFound ? id.errorPage.notFoundBody : id.errorPage.failedBody}
	</p>
	<a
		href="/"
		class="inline-flex min-h-12 items-center justify-center self-start rounded-control bg-accent px-5 font-semibold text-on-accent"
	>
		{id.errorPage.backHome}
	</a>
</main>
