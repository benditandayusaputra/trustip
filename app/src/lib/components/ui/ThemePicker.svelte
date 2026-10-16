<script lang="ts">
	import { onMount } from 'svelte';
	import SunMoon from '@lucide/svelte/icons/sun-moon';
	import { appearance as text } from '#lib/i18n/appearance.js';

	onMount(() => {
		window.dispatchEvent(new Event('trustip:refresh-appearance'));
	});
</script>

<label
	class="inline-flex min-h-11 items-center gap-2 rounded-pill border border-border bg-bg px-2 text-sm text-text"
>
	<SunMoon class="hidden size-4 shrink-0 md:block" aria-hidden="true" />
	<span class="sr-only">{text.label}</span>
	<select data-appearance-picker class="min-h-11 max-w-full bg-transparent text-sm text-text">
		{#each text.choices as choice (choice.value)}
			<option value={choice.value}>{choice.label}</option>
		{/each}
	</select>
</label>
