<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Tab {
		key: string;
		label: string;
	}

	interface Props {
		label: string;
		tabs: Tab[];
		selected?: string;
		centered?: boolean;
		panel: Snippet<[string]>;
	}

	let {
		label,
		tabs,
		selected = $bindable(tabs[0]?.key ?? ''),
		centered = false,
		panel
	}: Props = $props();

	const baseId = $props.id();
	const stepByKey: Record<string, number> = { ArrowRight: 1, ArrowLeft: -1 };

	function moveSelection(event: KeyboardEvent) {
		const step = stepByKey[event.key];
		if (step === undefined) {
			return;
		}
		const current = tabs.findIndex((tab) => tab.key === selected);
		const next = tabs[(current + step + tabs.length) % tabs.length];
		if (next) {
			selected = next.key;
			document.getElementById(`${baseId}-tab-${next.key}`)?.focus();
		}
	}
</script>

<div class="flex flex-col gap-4">
	<div
		role="tablist"
		aria-label={label}
		tabindex="-1"
		class="segmented-tabs flex gap-1 overflow-x-auto {centered ? 'self-center' : ''}"
		onkeydown={moveSelection}
	>
		{#each tabs as tab (tab.key)}
			<button
				type="button"
				role="tab"
				id="{baseId}-tab-{tab.key}"
				aria-selected={tab.key === selected}
				aria-controls="{baseId}-panel"
				tabindex={tab.key === selected ? 0 : -1}
				onclick={() => (selected = tab.key)}
				class="min-h-11 border-b-2 px-4 text-base font-medium {tab.key === selected
					? 'border-accent text-text'
					: 'border-transparent text-text-muted'}"
			>
				{tab.label}
			</button>
		{/each}
	</div>
	<div id="{baseId}-panel" role="tabpanel" aria-labelledby="{baseId}-tab-{selected}" tabindex="0">
		{@render panel(selected)}
	</div>
</div>
