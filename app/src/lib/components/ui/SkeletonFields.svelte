<script lang="ts">
	import Skeleton from './Skeleton.svelte';
	let { rows = 4 }: { rows?: number } = $props();
</script>

<div class="flex min-w-0 flex-col gap-5">
	{#each { length: rows }, index (index)}
		<div class="flex flex-col gap-2">
			<div class="w-1/3"><Skeleton /></div>
			<Skeleton shape="control" />
		</div>
	{/each}
	<div class="w-1/2"><Skeleton shape="control" /></div>
</div>
