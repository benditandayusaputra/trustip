<script lang="ts">
	import FileQuestion from '@lucide/svelte/icons/file-question';
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import { navigating, page } from '$app/state';
	import NavigationContent from '#lib/components/layout/NavigationContent.svelte';
	import { pendingPagePath } from '#lib/components/layout/page-skeleton.js';
	import { common as id } from '#lib/i18n/common.js';
	import ThemePicker from '#lib/components/ui/ThemePicker.svelte';

	const notFoundStatus = 404;
	const isNotFound = $derived(page.status === notFoundStatus);
</script>

<svelte:head>
	<title>{isNotFound ? id.errorPage.notFoundTitle : id.errorPage.failedTitle}</title>
</svelte:head>

<main id="konten" class="mx-auto flex min-h-dvh max-w-form flex-col justify-center gap-4 px-4">
	<NavigationContent pendingPath={pendingPagePath(navigating.to?.url, page.url)}>
		<div class="self-end"><ThemePicker /></div>
		<FileQuestion class="size-12 text-accent" aria-hidden="true" />
		<p class="font-mono text-kpi font-bold text-accent">{page.status}</p>
		<h1 class="text-xl">{isNotFound ? id.errorPage.notFoundTitle : id.errorPage.failedTitle}</h1>
		<p class="text-text-muted">
			{isNotFound ? id.errorPage.notFoundBody : id.errorPage.failedBody}
		</p>
		<a
			href="/"
			class="inline-flex min-h-12 items-center justify-center gap-2 self-start rounded-control bg-primary px-5 font-semibold text-on-accent"
		>
			<ArrowLeft class="size-5" aria-hidden="true" />{id.errorPage.backHome}
		</a>
	</NavigationContent>
</main>
