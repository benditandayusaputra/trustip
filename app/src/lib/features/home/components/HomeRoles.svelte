<script lang="ts">
	import ChefHat from '@lucide/svelte/icons/chef-hat';
	import Store from '@lucide/svelte/icons/store';
	import Check from '@lucide/svelte/icons/check';
	import ArrowRight from '@lucide/svelte/icons/arrow-right';
	import { home as id } from '#lib/i18n/home.js';

	const text = id.home;
	const icons = [ChefHat, Store];
</script>

<section class="flex flex-col gap-6" aria-labelledby="beranda-peran">
	<header class="flex flex-col gap-2">
		<p class="section-label">{text.rolesEyebrow}</p>
		<h2 id="beranda-peran" class="section-title">{text.rolesTitle}</h2>
		<p class="max-w-reading text-base text-text-muted">{text.rolesIntro} {text.audienceDetail}</p>
	</header>
	<div class="grid gap-4 md:grid-cols-2">
		{#each text.roles as role, index (role.title)}
			{@const Icon = icons[index] ?? Store}
			<article class="surface-card flex flex-col items-start gap-5 p-5 md:p-6">
				<span class="flex size-12 items-center justify-center rounded-control bg-accent-surface"
					><Icon class="size-6 text-accent" aria-hidden="true" /></span
				>
				<div class="flex flex-col gap-2">
					<h3 class="text-xl">{role.title}</h3>
					<p class="text-base text-text-muted">{role.body}</p>
				</div>
				<ul class="flex flex-col gap-2 text-sm text-text-muted">
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
