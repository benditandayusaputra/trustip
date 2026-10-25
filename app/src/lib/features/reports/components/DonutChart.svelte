<script lang="ts" module>
	import type { ChartTone } from '#lib/components/ui/tone.js';

	export interface ChartSlice {
		key: string;
		label: string;
		value: number;
		valueLabel: string;
		shareLabel: string;
		tone: ChartTone;
	}
</script>

<script lang="ts">
	import { donutSlices } from '#lib/domain/chart.js';
	import { chartDot, chartStroke } from '#lib/components/ui/tone.js';

	interface Props {
		caption: string;
		slices: ChartSlice[];
		totalLabel: string;
		totalValue: string;
	}

	let { caption, slices, totalLabel, totalValue }: Props = $props();

	const size = 176;
	const thickness = 22;
	const half = 0.5;
	const labelLift = 8;
	const valueDrop = 14;
	const centre = size * half;
	const radius = (size - thickness) * half;
	const circumference = Math.PI * (size - thickness);
	const arcs = $derived(
		donutSlices(
			slices.map((slice) => slice.value),
			circumference
		)
	);
</script>

<figure class="flex flex-col gap-4">
	<figcaption class="text-sm text-text-muted">{caption}</figcaption>
	<div class="flex flex-col items-center gap-6">
		<svg viewBox="0 0 {size} {size}" width={size} height={size} aria-hidden="true" class="shrink-0">
			<circle
				cx={centre}
				cy={centre}
				r={radius}
				fill="none"
				stroke-width={thickness}
				class="stroke-surface-hover"
			/>
			{#each slices as slice, index (slice.key)}
				{@const arc = arcs[index]}
				{#if arc && arc.length > 0}
					<circle
						cx={centre}
						cy={centre}
						r={radius}
						fill="none"
						stroke-width={thickness}
						stroke-dasharray="{arc.length} {circumference}"
						stroke-dashoffset={arc.offset}
						transform="rotate(-90 {centre} {centre})"
						class="chart-mark {chartStroke[slice.tone]}"
						><title>{slice.label}: {slice.valueLabel}</title></circle
					>
				{/if}
			{/each}
			<text x={centre} y={centre - labelLift} text-anchor="middle" class="fill-text-muted text-xs"
				>{totalLabel}</text
			>
			<text
				x={centre}
				y={centre + valueDrop}
				text-anchor="middle"
				class="fill-text font-mono text-sm font-bold">{totalValue}</text
			>
		</svg>
		<ul class="grid w-full grid-cols-2 gap-4">
			{#each slices as slice (slice.key)}
				<li class="flex items-start gap-2">
					<span class="mt-1 size-4 shrink-0 rounded-pill {chartDot[slice.tone]}"></span>
					<span class="flex min-w-0 flex-col gap-1">
						<span class="text-sm text-text-muted">{slice.label}</span>
						<span class="font-mono font-bold text-text tabular-nums">{slice.valueLabel}</span>
						<span class="text-xs text-text-muted">{slice.shareLabel}</span>
					</span>
				</li>
			{/each}
		</ul>
	</div>
</figure>
