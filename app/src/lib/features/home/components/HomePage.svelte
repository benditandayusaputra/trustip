<script lang="ts">
	import ClipboardList from '@lucide/svelte/icons/clipboard-list';
	import Handshake from '@lucide/svelte/icons/handshake';
	import ShieldCheck from '@lucide/svelte/icons/shield-check';
	import Store from '@lucide/svelte/icons/store';
	import ArrowRight from '@lucide/svelte/icons/arrow-right';
	import { home as id } from '#lib/i18n/home.js';
	import TradeExample from './TradeExample.svelte';

	const text = id.home;
	const stepIcons = [Store, Handshake, ClipboardList, ShieldCheck];
	const linkClass = 'inline-flex min-h-12 items-center rounded-control px-5 font-semibold';
</script>

<svelte:head>
	<title>{text.title}</title>
	<meta name="description" content={text.intro} />
</svelte:head>

<div class="flex flex-col gap-12 text-text">
	<section class="grid items-center gap-8 py-6 lg:grid-cols-2" aria-labelledby="beranda-judul">
		<div class="flex flex-col gap-6">
			<p class="section-label">{text.audience}</p>
			<h1 id="beranda-judul" class="text-hero-compact font-semibold md:text-hero">
				{text.headline}<br /><span class="text-accent">{text.headlineAccent}</span>
			</h1>
			<p class="max-w-form text-base leading-body text-text-muted">{text.intro}</p>
			<div class="flex flex-wrap gap-2">
				<a href="/daftar" class="action-link"
					>{text.register}<ArrowRight class="size-5" aria-hidden="true" /></a
				>
				<a
					href="/masuk#coba-demo"
					class="{linkClass} border border-border-strong bg-surface text-text">{text.tryDemo}</a
				>
			</div>
			<p class="text-sm text-text-muted">{text.audienceDetail}</p>
		</div>
		<TradeExample />
	</section>

	<section class="flex flex-col gap-4" aria-labelledby="beranda-masalah">
		<h2 id="beranda-masalah" class="text-lg font-semibold">{text.problemsTitle}</h2>
		<ul class="grid gap-4 md:grid-cols-2">
			{#each text.problems as problem (problem.title)}
				<li class="flex flex-col gap-2 rounded-card border border-border bg-surface p-4">
					<h3 class="text-base font-semibold">{problem.title}</h3>
					<p class="text-base text-text-muted">{problem.body}</p>
				</li>
			{/each}
		</ul>
	</section>

	<section class="flex flex-col gap-4" aria-labelledby="beranda-cara">
		<h2 id="beranda-cara" class="text-lg font-semibold">{text.stepsTitle}</h2>
		<ol class="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
			{#each text.steps as step, index (step.title)}
				{@const Icon = stepIcons[index] ?? Store}
				<li class="flex flex-col gap-4 border-t border-border pt-6">
					<Icon class="size-6 shrink-0 text-accent" aria-hidden="true" />
					<div class="flex flex-col gap-2">
						<h3 class="text-base font-semibold">{step.title}</h3>
						<p class="text-base text-text-muted">{step.body}</p>
					</div>
				</li>
			{/each}
		</ol>
	</section>

	<section class="trade-panel flex flex-col gap-4" aria-labelledby="beranda-mulai">
		<h2 id="beranda-mulai" class="text-lg font-semibold">{text.closingTitle}</h2>
		<p class="text-base text-text-muted">{text.closingBody}</p>
		<div class="flex flex-wrap gap-4">
			<a href="/daftar" class="action-link">{text.register}</a>
			<a href="/cari/toko" class="{linkClass} text-accent">{text.browseStores}</a>
			<a href="/kebutuhan" class="{linkClass} text-accent">{text.browseNeeds}</a>
		</div>
	</section>
</div>
