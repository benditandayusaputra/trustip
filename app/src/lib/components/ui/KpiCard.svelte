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
	}

	let {
		label,
		value,
		caption,
		icon,
		tone = 'info',
		loading = false,
		fullOnMobile = false
	}: Props = $props();
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
	class="flex min-w-0 flex-col gap-4 rounded-card border border-border bg-surface p-4 compact:p-5 {mobileSpan}"
>
	<div class="flex items-start justify-between gap-2">
		<p class="text-sm font-medium text-text-muted">{label}</p>
		{#if icon}
			<span
				class="flex size-11 shrink-0 items-center justify-center rounded-control {badgeTone[tone]}"
			>
				{@render icon()}
			</span>
		{/if}
	</div>
	{#if loading}
		<Skeleton shape="metric" />
	{:else}
		<p
			class="min-h-8 font-mono leading-tight font-bold wrap-anywhere text-text tabular-nums {valueSize}"
		>
			{value}
		</p>
	{/if}
	{#if caption}
		<p class="mt-auto border-t border-border pt-2 text-xs text-text-muted">{caption}</p>
	{/if}
</div>
