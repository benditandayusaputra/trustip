<script lang="ts">
	import Skeleton from './Skeleton.svelte';
	let { rows = 4 }: { rows?: number } = $props();
</script>

<div class="grid grid-cols-(--product-grid-columns) gap-5">
	{#each { length: rows }, index (index)}
		<div class="surface-card flex min-w-0 flex-col gap-4 p-4">
			<Skeleton shape="image" />
			<div class="w-1/2"><Skeleton /></div>
			<Skeleton shape="metric" />
			<div class="w-1/3"><Skeleton /></div>
			<div class="border-t border-border pt-4"><Skeleton shape="metric" /></div>
		</div>
	{/each}
</div>
