<script lang="ts">
	interface Shape {
		line: string;
		area: string;
	}

	interface Props {
		caption: string;
		shape: Shape;
		width: number;
		height: number;
		ceilingLabel: string;
		firstLabel: string;
		lastLabel: string;
	}

	let { caption, shape, width, height, ceilingLabel, firstLabel, lastLabel }: Props = $props();

	const gridRows = 3;
	const gradientId = $props.id();
	const gridLines = $derived(
		Array.from({ length: gridRows }, (_, index) => ((index + 1) * height) / (gridRows + 1))
	);
</script>

<figure class="flex flex-col gap-2">
	<figcaption class="text-sm text-text-muted">{caption}</figcaption>
	<p class="font-mono text-xs text-text-muted tabular-nums">{ceilingLabel}</p>
	<svg
		viewBox="0 0 {width} {height}"
		preserveAspectRatio="none"
		role="img"
		aria-label={caption}
		{height}
		class="w-full"
	>
		<defs>
			<linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
				<stop offset="0" stop-color="var(--color-accent)" stop-opacity="0.24" />
				<stop offset="1" stop-color="var(--color-chart-fill)" stop-opacity="0" />
			</linearGradient>
		</defs>
		{#each gridLines as top (top)}
			<line
				x1="0"
				x2={width}
				y1={top}
				y2={top}
				stroke="var(--color-border)"
				stroke-width="1"
				vector-effect="non-scaling-stroke"
			/>
		{/each}
		<path d={shape.area} fill="url(#{gradientId})" />
		<path
			d={shape.line}
			fill="none"
			stroke="var(--color-accent)"
			stroke-width="2"
			vector-effect="non-scaling-stroke"
		/>
	</svg>
	<div class="flex justify-between text-xs text-text-muted">
		<span>{firstLabel}</span>
		<span>{lastLabel}</span>
	</div>
</figure>
