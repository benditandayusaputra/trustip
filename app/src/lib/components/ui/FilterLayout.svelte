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
	<div class="sticky-filters">
		<div class="filter-toolbar flex max-w-(--search-panel-width) flex-col gap-2">
			<div class="flex flex-wrap items-end gap-2">
				<div class="min-w-0 flex-1 basis-full md:basis-0">{@render searchControls()}</div>
				<button
					type="button"
					aria-expanded={open}
					aria-controls={panelId}
					onclick={() => (open = !open)}
					class="inline-flex min-h-11 items-center gap-2 rounded-control border border-border-strong bg-surface px-4 text-sm font-semibold text-text hover:bg-surface-hover"
				>
					<SlidersHorizontal class="size-4" aria-hidden="true" />{text.layout.filters}
					<ChevronDown class="size-4 {open ? 'rotate-180' : ''}" aria-hidden="true" />
				</button>
			</div>
			<div id={panelId} class="border-t border-border pt-4 {open ? '' : 'hidden'}">
				{@render filters()}
			</div>
		</div>
	</div>
	<div class="flex min-w-0 flex-col gap-4">{@render children()}</div>
</div>
