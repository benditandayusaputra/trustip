<script lang="ts">
	import UserRound from '@lucide/svelte/icons/user-round';
	import Search from '@lucide/svelte/icons/search';
	import Handshake from '@lucide/svelte/icons/handshake';
	import Package from '@lucide/svelte/icons/package';
	import Calculator from '@lucide/svelte/icons/calculator';
	import Wallet from '@lucide/svelte/icons/wallet';
	import ChartColumn from '@lucide/svelte/icons/chart-column';
	import ShieldCheck from '@lucide/svelte/icons/shield-check';
	import CloudUpload from '@lucide/svelte/icons/cloud-upload';
	import { common } from '#lib/i18n/common.js';
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
		ShieldCheck,
		CloudUpload
	];
	const personas = Object.values(common.auth.login.demo.accounts);
</script>

<article class="flex flex-col gap-8">
	<header class="grid items-start gap-8 lg:grid-cols-12">
		<div class="flex flex-col gap-4 lg:col-span-7">
			<p class="section-label">{text.eyebrow}</p>
			<h1 class="section-title text-text">{text.headline}</h1>
			<p class="reading-text text-base text-text-muted">{text.intro}</p>
			<div class="flex flex-wrap gap-4">
				<a href="/demo" class="action-link">{text.tryDemo}</a>
				<a href="/daftar" class="secondary-link">{text.register}</a>
			</div>
		</div>
		<aside class="trade-panel flex flex-col gap-4 lg:col-span-5" aria-labelledby="panduan-akun">
			<h2 id="panduan-akun" class="text-lg font-semibold text-text">{text.personasTitle}</h2>
			<p class="text-sm text-text-muted">{text.personasIntro}</p>
			<ul class="flex flex-col gap-4">
				{#each personas as persona (persona.name)}
					<li class="flex items-center gap-4">
						<span
							class="flex size-11 shrink-0 items-center justify-center rounded-pill bg-accent-surface font-semibold text-accent"
							aria-hidden="true">{persona.initial}</span
						>
						<span class="flex min-w-0 flex-col">
							<span class="font-semibold text-text">{persona.name}</span>
							<span class="text-sm text-text-muted">{persona.role}</span>
						</span>
					</li>
				{/each}
			</ul>
		</aside>
	</header>
	<div class="grid items-start gap-8 lg:grid-cols-12">
		<nav
			aria-label={text.contentsLabel}
			class="trade-panel flex flex-col gap-2 lg:sticky lg:top-(--header-height) lg:col-span-3"
		>
			<p class="text-sm font-semibold text-text">{text.contents}</p>
			<ol class="flex flex-col">
				{#each text.phases as phase, index (phase.slug)}
					<li>
						<a
							href="#{phase.slug}"
							class="flex min-h-11 items-baseline gap-2 py-2 text-sm text-text-muted hover:text-accent"
							><span class="font-mono text-xs">{index + 1}</span>{phase.title}</a
						>
					</li>
				{/each}
			</ol>
		</nav>
		<ol class="flex min-w-0 flex-col gap-6 lg:col-span-9">
			{#each text.phases as phase, index (phase.slug)}
				<GuidePhaseCard {phase} number={index + 1} icon={icons[index] ?? UserRound} />
			{/each}
		</ol>
	</div>
	<section class="trade-panel flex flex-col gap-4" aria-labelledby="panduan-skenario">
		<h2 id="panduan-skenario" class="text-lg font-semibold text-text">{text.scenarioTitle}</h2>
		<p class="text-base text-text-muted">{text.scenarioIntro}</p>
		<ol class="flex list-decimal flex-col gap-2 pl-5 text-base text-text">
			{#each text.scenario as step (step)}
				<li class="pl-1">{step}</li>
			{/each}
		</ol>
	</section>
	<section
		class="flex flex-col items-start gap-4 rounded-card bg-accent-surface p-6"
		aria-labelledby="panduan-faq"
	>
		<h2 id="panduan-faq" class="text-lg font-semibold text-text">{text.faqTitle}</h2>
		<p class="text-base text-text-muted">{text.faqBody}</p>
		<a href="/faq" class="action-link">{text.faqLink}</a>
	</section>
</article>
