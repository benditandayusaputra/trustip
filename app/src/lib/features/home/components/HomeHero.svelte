<script lang="ts">
	import ArrowRight from '@lucide/svelte/icons/arrow-right';
	import ArrowDown from '@lucide/svelte/icons/arrow-down';
	import Check from '@lucide/svelte/icons/check';
	import ChefHat from '@lucide/svelte/icons/chef-hat';
	import Store from '@lucide/svelte/icons/store';
	import TradeIllustration from '#lib/components/ui/TradeIllustration.svelte';
	import { home as id } from '#lib/i18n/home.js';

	const text = id.home;
</script>

<section
	class="grid items-center gap-8 py-6 lg:grid-cols-2 lg:py-8"
	aria-labelledby="beranda-judul"
>
	<div class="flex flex-col items-start gap-6">
		<p
			class="inline-flex items-center gap-2 rounded-pill border border-border bg-surface px-4 py-2 text-xs font-semibold text-accent"
		>
			<span class="size-2 rounded-pill bg-primary" aria-hidden="true"></span>{text.audience}
		</p>
		<h1 id="beranda-judul" class="text-hero-compact font-semibold md:text-hero">
			{text.headline}<br /><span class="text-accent">{text.headlineAccent}</span>
		</h1>
		<p class="max-w-form text-base text-text-muted">{text.intro}</p>
		<div class="flex flex-wrap gap-2">
			<a href="/daftar" class="action-link"
				>{text.register}<ArrowRight class="size-5" aria-hidden="true" /></a
			>
			<a
				href="/demo"
				class="inline-flex min-h-12 items-center justify-center rounded-control border border-border-strong bg-surface px-5 font-semibold text-text"
				>{text.tryDemo}</a
			>
		</div>
		<ul class="flex flex-wrap gap-x-4 gap-y-2 text-sm text-text-muted">
			{#each text.highlights as highlight (highlight)}
				<li class="flex items-center gap-2">
					<Check class="size-4 shrink-0 text-accent" aria-hidden="true" />{highlight}
				</li>
			{/each}
		</ul>
		<a
			href="#beranda-cara"
			class="inline-flex min-h-11 items-center gap-2 font-semibold text-accent hover:underline"
			>{text.seeWorkflow}<ArrowDown class="size-4" aria-hidden="true" /></a
		>
	</div>
	<figure class="hero-art flex flex-col gap-4 border border-border p-5 md:p-6">
		<p class="section-label">{text.illustration.eyebrow}</p>
		<TradeIllustration />
		<figcaption class="flex flex-col gap-4">
			<p class="text-center text-lg font-semibold text-text">{text.illustration.caption}</p>
			<div class="grid grid-cols-2 gap-4">
				<div class="flex flex-col gap-2 rounded-control bg-surface p-4">
					<ChefHat class="size-5 text-accent" aria-hidden="true" />
					<p class="font-semibold text-text">{text.illustration.consignor}</p>
					<p class="text-sm text-text-muted">{text.illustration.consignorBody}</p>
				</div>
				<div class="flex flex-col gap-2 rounded-control bg-surface p-4">
					<Store class="size-5 text-accent" aria-hidden="true" />
					<p class="font-semibold text-text">{text.illustration.store}</p>
					<p class="text-sm text-text-muted">{text.illustration.storeBody}</p>
				</div>
			</div>
		</figcaption>
	</figure>
</section>
