<script lang="ts">
	import Skeleton from './Skeleton.svelte';
	let { rows = 4 }: { rows?: number } = $props();
</script>

<div class="flex min-w-0 flex-col gap-2">
	{#each { length: rows }, index (index)}
		<div
			class="flex min-w-0 items-center gap-4 rounded-control border border-border bg-surface p-4"
		>
			<div class="size-12 shrink-0"><Skeleton shape="square" /></div>
			<div class="flex min-w-0 flex-1 flex-col gap-2">
				<div class="w-2/3"><Skeleton /></div>
				<div class="w-1/2"><Skeleton /></div>
			</div>
			<div class="w-1/4"><Skeleton shape="metric" /></div>
		</div>
	{/each}
</div>
