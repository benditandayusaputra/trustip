<script lang="ts">
	import type { Snippet } from 'svelte';
	import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import { common as text } from '#lib/i18n/common.js';
	let {
		filters,
		searchControls,
		children
	}: { filters: Snippet; searchControls: Snippet; children: Snippet } = $props();
	let open = $state(false);
	const panelId = $props.id();
</script>

<div class="flex min-w-0 flex-col gap-5">
	<div class="filter-toolbar">{@render searchControls()}</div>
	<div class="grid items-start gap-6 xl:grid-cols-(--directory-columns)">
		<aside class="flex min-w-0 flex-col gap-4">
			<button
				type="button"
				aria-expanded={open}
				aria-controls={panelId}
				onclick={() => (open = !open)}
				class="inline-flex min-h-12 items-center justify-between gap-2 rounded-control border border-border-strong bg-surface px-4 text-sm font-semibold xl:hidden"
			>
				<span class="flex items-center gap-2"
					><SlidersHorizontal class="size-5" aria-hidden="true" />{text.layout.filters}</span
				><ChevronDown class="size-4" aria-hidden="true" />
			</button>
			<div id={panelId} class="trade-panel {open ? '' : 'hidden xl:block'}">
				{@render filters()}
			</div>
		</aside>
		<div class="flex min-w-0 flex-col gap-4">{@render children()}</div>
	</div>
</div>
