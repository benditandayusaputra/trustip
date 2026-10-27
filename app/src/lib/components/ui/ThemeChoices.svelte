<script lang="ts">
	import Check from '@lucide/svelte/icons/check';
	import Moon from '@lucide/svelte/icons/moon';
	import Sun from '@lucide/svelte/icons/sun';
	import { appearance as text } from '#lib/i18n/appearance.js';

	const icons = { light: Sun, dark: Moon };
</script>

{#each text.choices as choice (choice.value)}
	{@const Icon = icons[choice.value]}
	<button
		type="button"
		data-appearance-choice={choice.value}
		aria-pressed="false"
		class="theme-choice inline-flex min-h-11 items-center gap-2 rounded-control px-4 text-sm font-medium"
	>
		<Icon class="size-4 shrink-0" aria-hidden="true" />{choice.label}
		<Check class="theme-check ml-auto size-4 text-accent" aria-hidden="true" />
	</button>
{/each}
