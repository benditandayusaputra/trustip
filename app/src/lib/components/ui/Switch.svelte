<script lang="ts">
	interface Props {
		label: string;
		checked?: boolean;
		description?: string;
		busy?: boolean;
		onchange?: (checked: boolean) => void;
	}

	let { label, checked = $bindable(false), description, busy = false, onchange }: Props = $props();

	const switchId = $props.id();
</script>

<div class="flex min-h-11 items-center justify-between gap-4">
	<div class="flex flex-col">
		<span id="{switchId}-label" class="text-base text-text">{label}</span>
		{#if description}
			<span id="{switchId}-description" class="text-xs text-text-muted">{description}</span>
		{/if}
	</div>
	<button
		type="button"
		role="switch"
		aria-checked={checked}
		aria-labelledby="{switchId}-label"
		aria-describedby={description ? `${switchId}-description` : undefined}
		disabled={busy}
		aria-busy={busy}
		onclick={() => {
			checked = !checked;
			onchange?.(checked);
		}}
		class="relative inline-flex h-8 w-12 shrink-0 items-center rounded-pill border border-border-strong px-1 {checked
			? 'bg-accent'
			: 'bg-bg'}"
	>
		<span
			class="size-6 rounded-pill transition-transform duration-(--duration-fast) ease-out {checked
				? 'translate-x-4 bg-on-accent'
				: 'translate-x-0 bg-text-muted'}"
		></span>
	</button>
</div>
