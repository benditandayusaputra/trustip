<script lang="ts">
	import { home as id } from '#lib/i18n/home.js';

	const text = id.home;
</script>

<section
	id="beranda-cara"
	tabindex="-1"
	class="soft-panel flex flex-col gap-8 p-5 md:p-8 lg:p-12"
	aria-labelledby="beranda-cara-judul"
>
	<header class="flex flex-col gap-2">
		<h2 id="beranda-cara-judul" class="section-title font-bold">{text.stepsTitle}</h2>
		<p class="reading-text max-w-reading text-base text-text-muted md:text-lg">
			{text.stepsIntro}
		</p>
	</header>
	<ol class="grid gap-6 md:grid-cols-2">
		{#each text.steps as step, index (step.title)}
			<li class="surface-card flex flex-col gap-4 p-5 md:p-6 md:last:col-span-2">
				<div class="flex items-center gap-4">
					<span
						class="inline-flex size-11 shrink-0 items-center justify-center rounded-control bg-accent-surface text-lg font-bold text-accent"
						>{index + 1}</span
					>
					<div class="flex flex-col">
						<h3 class="text-lg font-semibold">{step.title}</h3>
						<p class="text-xs font-medium text-accent">{step.actor}</p>
					</div>
				</div>
				<p class="reading-text text-sm text-text-muted">{step.body}</p>
			</li>
		{/each}
	</ol>
	<p class="rounded-control border border-border bg-accent-surface p-4 text-sm text-text-muted">
		{text.paymentNote}
	</p>
</section>
