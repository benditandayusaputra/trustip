<script lang="ts">
	import { onMount } from 'svelte';
	import Moon from '@lucide/svelte/icons/moon';
	import Sun from '@lucide/svelte/icons/sun';
	import { appearance as text } from '#lib/i18n/appearance.js';

	onMount(() => {
		window.dispatchEvent(new Event('trustip:refresh-appearance'));
	});
</script>

<section class="surface-card flex items-center justify-between gap-4 px-5 py-4 compact:px-6">
	<h2 class="text-lg font-semibold text-text">{text.title}</h2>
	<label class="flex min-h-11 items-center gap-2">
		<span class="sr-only">{text.dark}</span>
		<Sun class="size-5 text-warning" aria-hidden="true" />
		<input type="checkbox" role="switch" data-appearance-toggle class="appearance-switch" />
		<Moon class="size-5 text-accent" aria-hidden="true" />
	</label>
</section>
