<script lang="ts">
	import type { Snippet } from 'svelte';
	import Skeleton from './Skeleton.svelte';
	import { badgeTone, type Tone } from './tone';
	interface Props {
		label: string;
		value: string;
		caption?: string;
		icon?: Snippet;
		tone?: Tone;
		loading?: boolean;
		fullOnMobile?: boolean;
		featured?: boolean;
	}

	let {
		label,
		value,
		caption,
		icon,
		tone = 'info',
		loading = false,
		fullOnMobile = false,
		featured = false
	}: Props = $props();
	const iconTone = $derived(featured ? 'bg-sidebar-badge text-sidebar-accent' : badgeTone[tone]);
	const maximumNarrowCharacters = 8;
	const maximumLargeCharacters = 10;
	const mobileSpan = $derived(
		fullOnMobile || value.length > maximumNarrowCharacters ? 'col-span-full compact:col-span-1' : ''
	);
	const valueSize = $derived(
		value.length > maximumLargeCharacters ? 'text-xl' : 'text-xl xl:text-kpi-compact'
	);
</script>

<div
	data-featured={featured}
	class="metric-card flex min-w-0 flex-col gap-4 rounded-card border {featured
		? 'border-featured bg-featured'
		: 'border-border bg-surface'} p-4 compact:p-5 {mobileSpan}"
>
	<div class="flex items-start justify-between gap-2">
		<p class="metric-label text-sm font-medium">{label}</p>
		{#if icon}
			<span class="metric-icon flex size-11 shrink-0 items-center justify-center {iconTone}">
				{@render icon()}
			</span>
		{/if}
	</div>
	{#if loading}
		<Skeleton shape="metric" />
	{:else}
		<p
			class="metric-value min-h-8 font-mono leading-tight font-bold wrap-anywhere tabular-nums {valueSize}"
		>
			{value}
		</p>
	{/if}
	{#if caption}
		<p class="metric-caption mt-auto border-t border-border pt-2 text-xs">
			{caption}
		</p>
	{/if}
</div>
