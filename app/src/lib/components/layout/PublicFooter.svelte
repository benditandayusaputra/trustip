<script lang="ts">
	import type { Snippet } from 'svelte';
	import Brand from './Brand.svelte';
	import { publicFooter as text } from '#lib/i18n/public-footer.js';

	let { note }: { note?: Snippet | undefined } = $props();
	const year = new Date().getFullYear();
</script>

<footer class="bg-bg">
	<div class="mx-auto flex w-full max-w-content flex-col gap-8 px-4 py-8 lg:px-8">
		<div class="grid gap-8 md:grid-cols-2 lg:grid-cols-4">
			<div class="flex flex-col items-start gap-4">
				<Brand />
				<p class="text-sm text-text-muted">{text.description}</p>
			</div>
			{#each text.groups as group (group.title)}
				<nav aria-label={group.label} class="flex flex-col gap-2">
					<h2 class="text-xs font-semibold tracking-(--letter-spacing-label) text-text uppercase">
						{group.title}
					</h2>
					<ul>
						{#each group.links as link (link.href)}
							<li>
								<a
									href={link.href}
									class="inline-flex min-h-11 items-center text-sm text-text-muted hover:text-accent hover:underline"
									>{link.label}</a
								>
							</li>
						{/each}
					</ul>
				</nav>
			{/each}
		</div>
		<div
			class="flex flex-col gap-2 pt-6 text-xs text-text-muted md:flex-row md:flex-wrap md:items-center md:justify-between"
		>
			<p>{text.copyright(year)}</p>
			<p>{text.note}</p>
			{@render note?.()}
		</div>
	</div>
</footer>
