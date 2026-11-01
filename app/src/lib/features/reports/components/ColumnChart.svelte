<script lang="ts" module>
	export interface ColumnPoint {
		key: string;
		label: string;
		valueLabel: string;
		tick: string;
		tickNote: string;
	}
</script>

<script lang="ts">
	import type { ColumnChartShape } from '#lib/domain/chart.js';

	interface Props {
		caption: string;
		shape: ColumnChartShape;
		points: ColumnPoint[];
		ceilingLabel: string;
	}

	let { caption, shape, points, ceilingLabel }: Props = $props();

	const minimumColumnWidth = 24;
	const columnRadius = 4;
	const gridLines = ['0%', '25%', '50%', '75%'];
</script>

<figure class="flex flex-1 flex-col gap-2">
	<figcaption class="text-sm text-text-muted">{caption}</figcaption>
	<p class="font-mono text-xs text-text-muted tabular-nums">{ceilingLabel}</p>
	<div class="flex flex-1 flex-col overflow-x-auto">
		<div
			class="flex flex-1 flex-col gap-1 pb-1"
			style="min-width: {points.length * minimumColumnWidth}px"
		>
			<svg
				width="100%"
				role="img"
				aria-label={caption}
				class="block min-h-(--chart-min-height) flex-1"
			>
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
			<div
				class="grid text-center text-xs leading-tight text-text-muted tabular-nums"
				style="grid-template-columns: repeat({points.length}, minmax(0, 1fr))"
				aria-hidden="true"
			>
				{#each points as point (point.key)}
					<span class="flex flex-col"
						><span>{point.tick}</span>{#if point.tickNote}<span>{point.tickNote}</span>{/if}</span
					>
				{/each}
			</div>
		</div>
	</div>
</figure>
