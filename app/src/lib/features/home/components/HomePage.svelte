<script lang="ts">
	import ArrowRight from '@lucide/svelte/icons/arrow-right';
	import Check from '@lucide/svelte/icons/check';
	import { home as id } from '#lib/i18n/home.js';
	import HomeHero from './HomeHero.svelte';
	import HomeRoles from './HomeRoles.svelte';
	import HomeWorkflow from './HomeWorkflow.svelte';
	import TradeExample from './TradeExample.svelte';

	const text = id.home;
</script>

<svelte:head>
	<title>{text.title}</title>
	<meta name="description" content={text.intro} />
</svelte:head>

<div class="home-page flex flex-col gap-12 text-text">
	<HomeHero />
	<HomeRoles />
	<HomeWorkflow />

	<section
		class="grid items-center gap-8 lg:grid-cols-(--home-example-columns)"
		aria-labelledby="beranda-contoh"
	>
		<div class="flex flex-col gap-6">
			<header class="flex flex-col gap-4">
				<p class="section-label">{text.exampleEyebrow}</p>
				<h2 id="beranda-contoh" class="text-hero-compact font-bold md:text-hero">
					{text.exampleTitle}
				</h2>
				<p class="reading-text max-w-reading text-base text-text-muted md:text-lg">
					{text.exampleIntro}
				</p>
			</header>
			<ul class="flex flex-col gap-4 text-base text-text">
				{#each text.exampleHighlights as highlight (highlight)}
					<li class="flex items-start gap-4">
						<Check class="size-5 shrink-0 text-accent" aria-hidden="true" />{highlight}
					</li>
				{/each}
			</ul>
		</div>
		<TradeExample />
	</section>

	<section class="flex flex-col gap-6" aria-labelledby="beranda-masalah">
		<header class="flex flex-col items-center gap-2 text-center">
			<h2 id="beranda-masalah" class="section-title font-bold">{text.problemsTitle}</h2>
			<p class="text-base text-text-muted md:text-lg">{text.problemsIntro}</p>
		</header>
		<ul class="grid gap-4 md:grid-cols-2">
			{#each text.problems as problem (problem.title)}
				<li class="soft-panel flex flex-col gap-4 p-5 md:p-6">
					<div class="flex flex-col gap-2">
						<h3 class="text-lg">{problem.title}</h3>
						<p class="reading-text text-sm text-text-muted">{problem.body}</p>
					</div>
					<p
						class="flex items-start gap-2 border-t border-border pt-4 text-sm font-semibold text-accent"
					>
						<Check class="size-5 shrink-0" aria-hidden="true" />{problem.solution}
					</p>
				</li>
			{/each}
		</ul>
	</section>

	<section
		class="flex flex-col items-start gap-6 rounded-card bg-primary p-6 text-on-accent shadow-raised md:px-12 md:py-8 lg:px-12"
		aria-labelledby="beranda-mulai"
	>
		<div class="flex flex-col gap-2">
			<h2 id="beranda-mulai" class="text-hero-compact font-bold md:text-hero">
				{text.closingTitle}
			</h2>
			<p class="max-w-reading text-base md:text-lg">{text.closingBody}</p>
		</div>
		<div class="flex flex-wrap items-center gap-6">
			<a
				href="/daftar"
				class="inline-flex min-h-12 items-center justify-center gap-2 rounded-control bg-on-accent px-5 font-semibold text-primary hover:underline"
				>{text.register}<ArrowRight class="size-5" aria-hidden="true" /></a
			>
			<a href="/cari/toko" class="inline-flex min-h-11 items-center font-semibold underline"
				>{text.browseStores}</a
			>
			<a href="/kebutuhan" class="inline-flex min-h-11 items-center font-semibold underline"
				>{text.browseNeeds}</a
			>
		</div>
	</section>
</div>
