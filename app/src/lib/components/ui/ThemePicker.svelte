<script lang="ts">
	import { onMount } from 'svelte';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import SunMoon from '@lucide/svelte/icons/sun-moon';
	import ThemeChoices from './ThemeChoices.svelte';
	import { appearance as text } from '#lib/i18n/appearance.js';

	let { expanded = false }: { expanded?: boolean } = $props();

	onMount(() => {
		window.dispatchEvent(new Event('trustip:refresh-appearance'));
	});
</script>

{#if expanded}
	<div
		role="group"
		aria-label={text.choicesLabel}
		class="theme-choices-inline grid grid-cols-2 gap-2"
	>
		<ThemeChoices />
	</div>
{:else}
	<details data-appearance-menu class="theme-menu relative">
		<summary
			aria-label={text.label}
			class="inline-flex min-h-11 min-w-11 items-center justify-center gap-2 rounded-pill border border-border bg-surface px-2 text-sm text-text hover:bg-surface-hover md:px-4"
		>
			<SunMoon class="size-5 shrink-0" aria-hidden="true" />
			<span class="hidden font-medium md:inline">{text.short}</span>
			<ChevronDown class="hidden size-4 text-text-muted md:block" aria-hidden="true" />
		</summary>
		<div
			role="group"
			aria-label={text.choicesLabel}
			class="absolute right-0 z-10 mt-2 flex w-max min-w-full flex-col gap-1 rounded-card border border-border bg-surface p-2 shadow-raised"
		>
			<ThemeChoices />
		</div>
	</details>
{/if}
