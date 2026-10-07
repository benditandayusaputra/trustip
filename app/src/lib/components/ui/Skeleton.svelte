<script lang="ts">
	import { id } from '#lib/i18n/id.js';

	interface Props {
		lines?: number;
		shape?: 'text' | 'control';
	}

	let { lines = 1, shape = 'text' }: Props = $props();

	const shapes = { text: 'h-4 rounded-control', control: 'h-12 rounded-control' };
</script>

<div class="flex flex-col gap-2" role="status" aria-label={id.feedback.loading}>
	{#each { length: lines }, index (index)}
		<div class="w-full animate-pulse bg-surface-hover {shapes[shape]}"></div>
	{/each}
</div>
