<script lang="ts">
	import { common as id } from '#lib/i18n/common.js';

	interface Props {
		lines?: number;
		shape?: 'text' | 'control' | 'chart' | 'metric';
	}

	let { lines = 1, shape = 'text' }: Props = $props();

	const shapes = {
		text: 'h-4 rounded-control',
		metric: 'h-8 rounded-control',
		control: 'h-12 rounded-control',
		chart: 'aspect-10/3 rounded-control'
	};
</script>

<div class="flex flex-col gap-2" role="status" aria-label={id.feedback.loading}>
	{#each { length: lines }, index (index)}
		<div class="w-full bg-surface-hover motion-safe:animate-pulse {shapes[shape]}"></div>
	{/each}
</div>
