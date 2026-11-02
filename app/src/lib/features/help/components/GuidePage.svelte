<script lang="ts">
	import UserRound from '@lucide/svelte/icons/user-round';
	import Search from '@lucide/svelte/icons/search';
	import Handshake from '@lucide/svelte/icons/handshake';
	import Package from '@lucide/svelte/icons/package';
	import Calculator from '@lucide/svelte/icons/calculator';
	import Wallet from '@lucide/svelte/icons/wallet';
	import ChartColumn from '@lucide/svelte/icons/chart-column';
	import CloudUpload from '@lucide/svelte/icons/cloud-upload';
	import { guide as text } from '#lib/i18n/guide.js';
	import GuidePhaseCard from './GuidePhaseCard.svelte';

	const icons = [
		UserRound,
		Search,
		Handshake,
		Package,
		Calculator,
		Wallet,
		ChartColumn,
		CloudUpload
	];
</script>

<article class="flex flex-col gap-8">
	<header class="flex max-w-reading flex-col gap-4">
		<p class="section-label">{text.eyebrow}</p>
		<h1 class="section-title text-text">{text.headline}</h1>
		<p class="text-base text-text-muted">{text.intro}</p>
		<div class="flex flex-wrap gap-4">
			<a href="/daftar" class="action-link">{text.register}</a>
		</div>
	</header>
	<ol class="grid items-start gap-4 md:grid-cols-2">
		{#each text.phases as phase, index (phase.slug)}
			<GuidePhaseCard {phase} number={index + 1} icon={icons[index] ?? UserRound} />
		{/each}
	</ol>
	<section
		class="flex flex-col items-start gap-4 rounded-card bg-accent-surface p-6"
		aria-labelledby="panduan-faq"
	>
		<h2 id="panduan-faq" class="text-lg font-semibold text-text">{text.faqTitle}</h2>
		<p class="text-base text-text-muted">{text.faqBody}</p>
		<a href="/faq" class="action-link">{text.faqLink}</a>
	</section>
</article>
