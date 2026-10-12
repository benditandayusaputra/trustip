<script lang="ts">
	import Skeleton from './Skeleton.svelte';
	let { rows = 4 }: { rows?: number } = $props();
	const previewLines = 2;
</script>

<div class="grid grid-cols-1 items-stretch gap-4 lg:grid-cols-2">
	{#each { length: rows }, index (index)}
		<div class="flex min-w-0 flex-col gap-4 rounded-card border border-border bg-surface p-5">
			<div class="flex items-center gap-4">
				<div class="size-12 shrink-0"><Skeleton shape="square" /></div>
				<div class="w-2/3"><Skeleton lines={previewLines} /></div>
			</div>
			<Skeleton lines={previewLines} />
			<div class="grid grid-cols-2 gap-4"><Skeleton /><Skeleton /></div>
			<Skeleton shape="control" />
		</div>
	{/each}
</div>
