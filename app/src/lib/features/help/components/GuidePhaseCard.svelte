<script lang="ts">
	import type { Component } from 'svelte';
	import type { GuidePhase } from '#lib/i18n/guide.js';
	import { guide as text } from '#lib/i18n/guide.js';

	interface Props {
		phase: GuidePhase;
		number: number;
		icon: Component<{ class?: string; 'aria-hidden'?: 'true' }>;
	}

	let { phase, number, icon: Icon }: Props = $props();
</script>

<li id={phase.slug} class="surface-card flex flex-col gap-5 p-5 compact:p-6">
	<header class="flex items-start gap-4">
		<span
			class="flex size-11 shrink-0 items-center justify-center rounded-pill bg-primary text-lg font-semibold text-on-accent"
			>{number}</span
		>
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<h2 class="text-lg font-semibold text-text">{phase.title}</h2>
			<p class="text-sm font-semibold text-accent">{phase.actor}</p>
		</div>
		<Icon class="size-6 shrink-0 text-text-muted" aria-hidden="true" />
	</header>
	<p class="text-base text-text-muted">{phase.summary}</p>
	<ol
		class="flex list-decimal flex-col gap-2 pl-5 text-base text-text"
		aria-label={text.stepsLabel}
	>
		{#each phase.steps as step (step)}
			<li class="pl-1">{step}</li>
		{/each}
	</ol>
	<div class="grid gap-4 md:grid-cols-2">
		<p class="rounded-control border border-border bg-surface-hover p-4 text-sm text-text-muted">
			<span class="block font-semibold text-text">{text.tipLabel}</span>{phase.tip}
		</p>
		<p class="rounded-control border border-accent-surface bg-accent-surface p-4 text-sm text-text">
			<span class="block font-semibold text-accent">{text.demoLabel}</span>{phase.demo}
		</p>
	</div>
</li>
