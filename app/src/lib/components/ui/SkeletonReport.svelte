<script lang="ts">
	import Skeleton from './Skeleton.svelte';
	import SkeletonRows from './SkeletonRows.svelte';
	let { rows = 4 }: { rows?: number } = $props();
	const metrics = 4;
</script>

<div class="flex min-w-0 flex-col gap-6">
	<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
		{#each { length: metrics }, index (index)}
			<div class="flex min-w-0 flex-col gap-4 rounded-card border border-border bg-surface p-4">
				<Skeleton /><Skeleton shape="metric" />
			</div>
		{/each}
	</div>
	<div class="flex flex-col gap-4 rounded-card border border-border bg-surface p-5">
		<div class="w-1/3"><Skeleton /></div>
		<Skeleton shape="chart" />
	</div>
	<SkeletonRows {rows} />
</div>
