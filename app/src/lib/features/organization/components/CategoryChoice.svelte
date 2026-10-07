<script lang="ts">
	import type { Category } from '#lib/api/organizations.js';

	interface Props {
		legend: string;
		categories: Category[];
		selected?: string[];
		error?: string | undefined;
	}

	let { legend, categories, selected = $bindable([]), error }: Props = $props();

	const groupId = $props.id();
</script>

<fieldset class="flex flex-col gap-2" aria-describedby={error ? `${groupId}-error` : undefined}>
	<legend class="mb-2 text-sm font-medium text-text">{legend}</legend>
	{#each categories as category (category.id)}
		<label class="flex min-h-11 items-center gap-4 text-base text-text">
			<input
				type="checkbox"
				value={category.id}
				bind:group={selected}
				class="size-6 shrink-0 accent-accent"
			/>
			{category.name}
		</label>
	{/each}
	{#if error}
		<p id="{groupId}-error" class="text-sm text-danger">{error}</p>
	{/if}
</fieldset>
