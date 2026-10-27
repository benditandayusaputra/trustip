<script lang="ts">
	import ArrowRight from '@lucide/svelte/icons/arrow-right';
	import Check from '@lucide/svelte/icons/check';
	import ChefHat from '@lucide/svelte/icons/chef-hat';
	import Store from '@lucide/svelte/icons/store';
	import TradeIllustration from '#lib/components/ui/TradeIllustration.svelte';
	import { home as id } from '#lib/i18n/home.js';

	const text = id.home;
	const parties = [
		{ icon: ChefHat, title: text.illustration.consignor, body: text.illustration.consignorBody },
		{ icon: Store, title: text.illustration.store, body: text.illustration.storeBody }
	];
</script>

<section
	class="grid items-center gap-8 py-6 lg:grid-cols-2 lg:py-8"
	aria-labelledby="beranda-judul"
>
	<div class="flex flex-col items-start gap-6">
		<p
			class="rounded-pill border border-border bg-surface px-4 py-2 text-xs font-semibold text-accent"
		>
			{text.audience}
		</p>
		<h1 id="beranda-judul" class="text-hero-compact font-bold md:text-hero">
			{text.headline}<br /><span class="text-accent">{text.headlineAccent}</span>
		</h1>
		<p class="reading-text max-w-form text-base text-text-muted md:text-lg">{text.intro}</p>
		<div class="flex flex-col gap-6">
			<div class="flex flex-wrap gap-4">
				<a href="/daftar" class="action-link"
					>{text.register}<ArrowRight class="size-5" aria-hidden="true" /></a
				>
				<a
					href="/demo"
					class="inline-flex min-h-12 items-center justify-center rounded-control border border-border-strong bg-surface px-6 font-semibold text-text hover:border-accent"
					>{text.tryDemo}</a
				>
			</div>
			<ul class="flex flex-wrap gap-x-6 gap-y-2 text-sm text-text-muted">
				{#each text.highlights as highlight (highlight)}
					<li class="flex items-center gap-2">
						<Check class="size-4 shrink-0 text-accent" aria-hidden="true" />{highlight}
					</li>
				{/each}
			</ul>
		</div>
	</div>
	<figure class="hero-art flex min-w-0 flex-col gap-5 border border-border p-4 md:p-8">
		<p class="section-label">{text.illustration.eyebrow}</p>
		<TradeIllustration />
		<figcaption class="grid grid-cols-1 gap-4 compact:grid-cols-2">
			<span class="sr-only">{text.illustration.caption}</span>
			{#each parties as party (party.title)}
				<span class="flex flex-col gap-1 rounded-control bg-surface p-4">
					<span class="flex items-center gap-2 font-semibold text-text">
						<party.icon class="size-5 shrink-0 text-accent" aria-hidden="true" />{party.title}
					</span>
					<span class="text-xs text-text-muted">{party.body}</span>
				</span>
			{/each}
		</figcaption>
	</figure>
</section>
