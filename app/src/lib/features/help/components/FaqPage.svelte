<script lang="ts">
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import { faq as text } from '#lib/i18n/faq.js';
</script>

<article class="mx-auto flex w-full max-w-reading flex-col gap-8">
	<header class="flex flex-col gap-4">
		<p class="section-label">{text.eyebrow}</p>
		<h1 class="section-title text-text">{text.title}</h1>
		<p class="text-base text-text-muted">{text.intro}</p>
		<nav aria-label={text.contentsLabel}>
			<ul class="flex flex-wrap gap-2">
				{#each text.groups as group (group.slug)}
					<li>
						<a
							href="#{group.slug}"
							class="inline-flex min-h-11 items-center rounded-pill border border-border bg-surface px-4 text-sm font-medium text-text hover:border-accent hover:text-accent"
							>{group.title}</a
						>
					</li>
				{/each}
			</ul>
		</nav>
	</header>
	{#each text.groups as group (group.slug)}
		<section id={group.slug} class="flex flex-col gap-4" aria-labelledby="{group.slug}-judul">
			<h2 id="{group.slug}-judul" class="text-lg font-semibold text-text">{group.title}</h2>
			<div class="flex flex-col gap-2">
				{#each group.items as item (item.question)}
					<details class="faq-item secondary-disclosure">
						<summary>
							<span>{item.question}</span>
							<ChevronDown class="faq-chevron size-5 shrink-0 text-text-muted" aria-hidden="true" />
						</summary>
						<p class="text-base text-text-muted">{item.answer}</p>
					</details>
				{/each}
			</div>
		</section>
	{/each}
	<section
		class="flex flex-col items-start gap-4 rounded-card bg-accent-surface p-6"
		aria-labelledby="faq-panduan"
	>
		<h2 id="faq-panduan" class="text-lg font-semibold text-text">{text.guideTitle}</h2>
		<p class="text-base text-text-muted">{text.guideBody}</p>
		<div class="flex flex-wrap gap-4">
			<a href="/panduan" class="action-link">{text.guideLink}</a>
			<a href="/demo" class="secondary-link">{text.tryDemo}</a>
		</div>
	</section>
</article>
