<script lang="ts">
	import ChefHat from '@lucide/svelte/icons/chef-hat';
	import Store from '@lucide/svelte/icons/store';
	import Check from '@lucide/svelte/icons/check';
	import ArrowRight from '@lucide/svelte/icons/arrow-right';
	import { home as id } from '#lib/i18n/home.js';

	const text = id.home;
	const icons = [ChefHat, Store];
</script>

<section class="flex flex-col gap-8 border-t border-border pt-12" aria-labelledby="beranda-peran">
	<header class="flex flex-col gap-2">
		<h2 id="beranda-peran" class="section-title font-bold">{text.rolesTitle}</h2>
		<p class="reading-text max-w-reading text-base text-text-muted md:text-lg">
			{text.rolesIntro}
			{text.audienceDetail}
		</p>
	</header>
	<div class="grid gap-6 md:grid-cols-2">
		{#each text.roles as role, index (role.title)}
			{@const Icon = icons[index] ?? Store}
			<article class="soft-panel flex flex-col items-start gap-5 p-6 md:p-8">
				<span
					class="flex size-12 items-center justify-center rounded-control bg-surface shadow-card"
					><Icon class="size-6 text-accent" aria-hidden="true" /></span
				>
				<div class="flex flex-col gap-2">
					<h3 class="text-xl font-semibold">{role.title}</h3>
					<p class="reading-text text-base text-text-muted">{role.body}</p>
				</div>
				<ul class="flex flex-col gap-2 text-sm text-text">
					{#each role.benefits as benefit (benefit)}
						<li class="flex items-start gap-2">
							<Check class="size-5 shrink-0 text-accent" aria-hidden="true" />{benefit}
						</li>
					{/each}
				</ul>
				<a
					href={role.href}
					class="mt-auto inline-flex min-h-11 items-center gap-2 font-semibold text-accent hover:underline"
					>{role.action}<ArrowRight class="size-4" aria-hidden="true" /></a
				>
			</article>
		{/each}
	</div>
</section>
