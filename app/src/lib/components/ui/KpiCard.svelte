<script lang="ts">
	interface Props {
		label: string;
		value: string;
		caption?: string;
	}

	let { label, value, caption }: Props = $props();
	const maximumNarrowCharacters = 8;
	const maximumLargeCharacters = 10;
	const mobileSpan = $derived(
		value.length > maximumNarrowCharacters ? 'col-span-full compact:col-span-1' : ''
	);
	const valueSize = $derived(
		value.length > maximumLargeCharacters ? 'text-xl' : 'text-xl xl:text-kpi-compact'
	);
</script>

<div
	class="flex min-w-0 flex-col gap-1 rounded-card border border-border bg-surface p-4 compact:p-5 {mobileSpan}"
>
	<p class="text-sm font-medium text-text-muted">{label}</p>
	<p class="font-mono leading-tight font-bold wrap-anywhere text-text tabular-nums {valueSize}">
		{value}
	</p>
	{#if caption}
		<p class="text-xs text-text-muted">{caption}</p>
	{/if}
</div>
