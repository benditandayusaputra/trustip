<script lang="ts">
	interface Option {
		value: string;
		label: string;
	}

	interface Props {
		legend: string;
		options: Option[];
		selected?: string[];
		error?: string | undefined;
	}

	let { legend, options, selected = $bindable([]), error }: Props = $props();

	const groupId = $props.id();
</script>

<fieldset class="flex flex-col gap-2" aria-describedby={error ? `${groupId}-error` : undefined}>
	<legend class="mb-2 text-sm font-medium text-text">{legend}</legend>
	{#each options as option (option.value)}
		<label class="flex min-h-11 items-center gap-4 text-base text-text">
			<input
				type="checkbox"
				value={option.value}
				bind:group={selected}
				class="size-6 shrink-0 accent-accent"
			/>
			{option.label}
		</label>
	{/each}
	{#if error}
		<p id="{groupId}-error" class="text-sm text-danger">{error}</p>
	{/if}
</fieldset>
