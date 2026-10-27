<script lang="ts">
	import { onMount } from 'svelte';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import SunMoon from '@lucide/svelte/icons/sun-moon';
	import { appearance as text } from '#lib/i18n/appearance.js';

	let { expanded = false }: { expanded?: boolean } = $props();

	onMount(() => {
		window.dispatchEvent(new Event('trustip:refresh-appearance'));
	});
</script>

<label
	title={text.label}
	class="theme-control {expanded
		? 'theme-control-expanded'
		: ''} inline-flex min-h-11 items-center gap-2 rounded-pill border border-border bg-surface px-2 text-sm text-text md:px-4"
>
	<SunMoon class="size-5 shrink-0" aria-hidden="true" />
	<span class="sr-only">{text.label}</span>
	{#if !expanded}
		<span class="hidden font-medium md:inline" aria-hidden="true">{text.short}</span>
		<ChevronDown class="hidden size-4 text-text-muted md:block" aria-hidden="true" />
	{/if}
	<select
		data-appearance-picker
		class="theme-select min-h-11 max-w-full bg-transparent text-sm text-text"
	>
		{#each text.choices as choice (choice.value)}
			<option value={choice.value}>{choice.label}</option>
		{/each}
	</select>
</label>
