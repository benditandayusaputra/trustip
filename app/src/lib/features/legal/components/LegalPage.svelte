<script lang="ts">
	import type { LegalDocument } from '#lib/i18n/legal.js';
	import { legal as legalText } from '#lib/i18n/legal.js';

	const id = { legal: legalText };

	interface Props {
		document: LegalDocument;
		related: { href: string; title: string };
	}

	let { document, related }: Props = $props();
</script>

<svelte:head>
	<title>{document.title}</title>
</svelte:head>

<article class="trade-panel mx-auto flex w-full max-w-reading flex-col gap-8 text-text">
	<header class="flex flex-col gap-2">
		<h1 class="text-xl font-semibold">{document.title}</h1>
		<p class="text-sm text-text-muted">{document.effective}</p>
		<p class="text-base">{document.intro}</p>
	</header>
	{#each document.sections as section (section.heading)}
		<section class="flex flex-col gap-2">
			<h2 class="text-lg font-semibold">{section.heading}</h2>
			<ul class="flex list-disc flex-col gap-2 pl-5 text-base">
				{#each section.points as point (point)}
					<li>{point}</li>
				{/each}
			</ul>
		</section>
	{/each}
	<p class="text-base text-text-muted">
		{id.legal.seeAlso}
		<a href={related.href} class="font-semibold text-accent">{related.title}</a>
	</p>
</article>
