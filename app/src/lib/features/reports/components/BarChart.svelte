<script lang="ts" module>
	import type { ChartTone } from '#lib/components/ui/tone.js';

	export interface BarSegment {
		label: string;
		value: number;
		valueLabel: string;
		tone: ChartTone;
	}

	export interface BarRow {
		key: string;
		label: string;
		valueLabel: string;
		segments: BarSegment[];
	}

	export interface LegendEntry {
		label: string;
		tone: ChartTone;
	}
</script>

<script lang="ts">
	import { stackedBars } from '#lib/domain/chart.js';
	import { chartDot, chartFill } from '#lib/components/ui/tone.js';

	interface Props {
		caption: string;
		rows: BarRow[];
		legend?: LegendEntry[];
	}

	let { caption, rows, legend = [] }: Props = $props();

	const barHeight = 12;
	const barRadius = 6;
	const clipId = $props.id();
	const layout = $derived(
		stackedBars(rows.map((row) => row.segments.map((segment) => segment.value)))
	);
</script>

<figure class="flex flex-col gap-4">
	<figcaption class="text-sm text-text-muted">{caption}</figcaption>
	{#if legend.length > 0}
		<ul class="flex flex-wrap gap-4 text-sm text-text-muted">
			{#each legend as entry (entry.label)}
				<li class="flex items-center gap-2">
					<span class="size-4 shrink-0 rounded-pill {chartDot[entry.tone]}"></span>{entry.label}
				</li>
			{/each}
		</ul>
	{/if}
	<ol class="flex flex-col gap-4">
		{#each rows as row, index (row.key)}
			{@const pieces = layout[index] ?? []}
			{@const reach = pieces.reduce((sum, piece) => sum + piece.width, 0)}
			<li class="flex flex-col gap-2">
				<p class="flex items-baseline justify-between gap-4 text-sm">
					<span class="min-w-0 font-medium wrap-anywhere text-text">{row.label}</span>
					<span class="shrink-0 text-right font-mono text-text tabular-nums">{row.valueLabel}</span>
				</p>
				<svg width="100%" height={barHeight} aria-hidden="true" class="block">
					<defs>
						<clipPath id="{clipId}-{index}">
							<rect width="{reach}%" height={barHeight} rx={barRadius} />
						</clipPath>
					</defs>
					<rect width="100%" height={barHeight} rx={barRadius} class="fill-surface-hover" />
					<g clip-path="url(#{clipId}-{index})">
						{#each row.segments as segment, part (segment.label)}
							{@const piece = pieces[part]}
							{#if piece}
								<rect
									x="{piece.start}%"
									width="{piece.width}%"
									height={barHeight}
									class="chart-mark {chartFill[segment.tone]}"
									><title>{segment.label}: {segment.valueLabel}</title></rect
								>
							{/if}
						{/each}
					</g>
				</svg>
			</li>
		{/each}
	</ol>
</figure>
