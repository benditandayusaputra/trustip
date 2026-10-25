<script lang="ts" module>
	export interface ColumnPoint {
		key: string;
		label: string;
		valueLabel: string;
	}
</script>

<script lang="ts">
	import type { ColumnChartShape } from '#lib/domain/chart.js';

	interface Props {
		caption: string;
		shape: ColumnChartShape;
		points: ColumnPoint[];
		ceilingLabel: string;
		firstLabel: string;
		lastLabel: string;
	}

	let { caption, shape, points, ceilingLabel, firstLabel, lastLabel }: Props = $props();

	const chartHeight = 192;
	const columnRadius = 4;
	const gridLines = ['0%', '25%', '50%', '75%'];
</script>

<figure class="flex flex-col gap-2">
	<figcaption class="text-sm text-text-muted">{caption}</figcaption>
	<p class="font-mono text-xs text-text-muted tabular-nums">{ceilingLabel}</p>
	<svg width="100%" height={chartHeight} role="img" aria-label={caption} class="block">
		{#each gridLines as line (line)}
			<line x1="0" x2="100%" y1={line} y2={line} stroke-width="1" class="stroke-border" />
		{/each}
		<line x1="0" x2="100%" y1="100%" y2="100%" stroke-width="2" class="stroke-border-strong" />
		{#each points as point, index (point.key)}
			{@const column = shape.columns[index]}
			{#if column && column.height > 0}
				<rect
					x="{column.left}%"
					y="{column.top}%"
					width="{column.width}%"
					height="{column.height}%"
					rx={columnRadius}
					class="chart-mark fill-primary"><title>{point.label}: {point.valueLabel}</title></rect
				>
			{/if}
		{/each}
	</svg>
	<div class="flex justify-between gap-4 text-xs text-text-muted">
		<span>{firstLabel}</span>
		<span>{lastLabel}</span>
	</div>
</figure>
