<script lang="ts">
	interface Props {
		label: string;
		value: string;
		caption?: string;
	}

	let { label, value, caption }: Props = $props();
</script>

<div class="flex flex-col gap-1 rounded-card border border-border bg-surface p-4 compact:p-5">
	<p class="text-sm font-medium text-text-muted">{label}</p>
	<p
		class="font-mono text-kpi-compact leading-tight font-bold text-accent tabular-nums compact:text-kpi"
	>
		{value}
	</p>
	{#if caption}
		<p class="text-xs text-text-muted">{caption}</p>
	{/if}
</div>
