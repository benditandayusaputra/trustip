<script lang="ts">
	import ClipboardList from '@lucide/svelte/icons/clipboard-list';
	import Handshake from '@lucide/svelte/icons/handshake';
	import ShieldCheck from '@lucide/svelte/icons/shield-check';
	import Store from '@lucide/svelte/icons/store';
	import { id } from '#lib/i18n/id.js';

	const text = id.home;
	const stepIcons = [Store, Handshake, ClipboardList, ShieldCheck];
	const linkClass = 'inline-flex min-h-12 items-center rounded-control px-5 font-semibold';
</script>

<svelte:head>
	<title>{text.title}</title>
	<meta name="description" content={text.intro} />
</svelte:head>

<div class="flex flex-col gap-12 text-text">
	<section class="flex flex-col gap-6 py-6" aria-labelledby="beranda-judul">
		<h1 id="beranda-judul" class="max-w-form text-xl font-semibold md:text-kpi">
			{text.headline}
		</h1>
		<p class="max-w-form text-base text-text-muted">{text.intro}</p>
		<div class="flex flex-wrap gap-4">
			<a href="/daftar" class="{linkClass} bg-accent text-on-accent">{text.register}</a>
			<a href="/masuk" class="{linkClass} border border-border-strong">{text.signIn}</a>
			<a href="/masuk#coba-demo" class="{linkClass} text-accent">{text.tryDemo}</a>
		</div>
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
		<ol class="grid gap-4 md:grid-cols-2">
			{#each text.steps as step, index (step.title)}
				{@const Icon = stepIcons[index] ?? Store}
				<li class="flex gap-4 rounded-card border border-border bg-surface p-4">
					<Icon class="size-6 shrink-0 text-accent" aria-hidden="true" />
					<div class="flex flex-col gap-2">
						<h3 class="text-base font-semibold">{step.title}</h3>
						<p class="text-base text-text-muted">{step.body}</p>
					</div>
				</li>
			{/each}
		</ol>
	</section>

	<section class="flex flex-col gap-4 pb-6" aria-labelledby="beranda-mulai">
		<h2 id="beranda-mulai" class="text-lg font-semibold">{text.closingTitle}</h2>
		<p class="text-base text-text-muted">{text.closingBody}</p>
		<div class="flex flex-wrap gap-4">
			<a href="/daftar" class="{linkClass} border border-border-strong">{text.register}</a>
			<a href="/cari/toko" class="{linkClass} text-accent">{text.browseStores}</a>
			<a href="/kebutuhan" class="{linkClass} text-accent">{text.browseNeeds}</a>
		</div>
	</section>
</div>
