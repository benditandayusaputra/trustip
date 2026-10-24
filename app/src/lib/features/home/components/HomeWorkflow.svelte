<script lang="ts">
	import Store from '@lucide/svelte/icons/store';
	import Handshake from '@lucide/svelte/icons/handshake';
	import Package from '@lucide/svelte/icons/package';
	import ClipboardList from '@lucide/svelte/icons/clipboard-list';
	import BadgeCheck from '@lucide/svelte/icons/badge-check';
	import { home as id } from '#lib/i18n/home.js';

	const text = id.home;
	const icons = [Store, Handshake, Package, ClipboardList, BadgeCheck];
</script>

<section
	id="beranda-cara"
	tabindex="-1"
	class="flex flex-col gap-6"
	aria-labelledby="beranda-cara-judul"
>
	<header class="flex flex-col gap-2">
		<p class="section-label">{text.stepsEyebrow}</p>
		<h2 id="beranda-cara-judul" class="section-title">{text.stepsTitle}</h2>
		<p class="max-w-reading text-base text-text-muted">{text.stepsIntro}</p>
	</header>
	<ol class="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
		{#each text.steps as step, index (step.title)}
			{@const Icon = icons[index] ?? Store}
			<li class="surface-card flex flex-col gap-4 p-5 md:last:col-span-2 lg:last:col-span-1">
				<div class="flex items-center justify-between gap-4">
					<span
						class="inline-flex size-11 items-center justify-center rounded-pill bg-accent-surface text-lg font-semibold text-accent"
						>{index + 1}</span
					>
					<Icon class="size-5 text-text-muted" aria-hidden="true" />
				</div>
				<div class="flex flex-col gap-2">
					<h3 class="text-lg">{step.title}</h3>
					<p class="text-xs font-semibold text-accent">{step.actor}</p>
					<p class="text-sm text-text-muted">{step.body}</p>
				</div>
			</li>
		{/each}
	</ol>
	<p class="rounded-control border border-border bg-accent-surface p-4 text-sm text-text-muted">
		{text.paymentNote}
	</p>
</section>
