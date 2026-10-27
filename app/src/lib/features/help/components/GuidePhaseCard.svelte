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

<li id={phase.slug} class="surface-card flex flex-col gap-2 p-5">
	<header class="flex items-start gap-2">
		<span
			class="flex size-8 shrink-0 items-center justify-center rounded-pill bg-primary text-sm font-semibold text-on-accent"
			>{number}</span
		>
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<h2 class="text-base font-semibold text-text">{phase.title}</h2>
			<p class="text-xs font-semibold text-accent">{phase.actor}</p>
		</div>
		<Icon class="size-5 shrink-0 text-text-muted" aria-hidden="true" />
	</header>
	<p class="text-sm text-text-muted">{phase.summary}</p>
	<details class="text-sm">
		<summary class="inline-flex min-h-11 cursor-pointer items-center font-semibold text-accent"
			><span>{text.stepsToggle}</span></summary
		>
		<ol class="flex list-decimal flex-col gap-2 pt-2 pl-5 text-text" aria-label={text.stepsLabel}>
			{#each phase.steps as step (step)}
				<li class="pl-1">{step}</li>
			{/each}
		</ol>
	</details>
</li>
