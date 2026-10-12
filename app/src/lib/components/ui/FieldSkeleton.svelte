<script lang="ts">
	import Skeleton from './Skeleton.svelte';
	let { label }: { label: string } = $props();
</script>

<div class="flex flex-col gap-2" aria-busy="true">
	<p class="text-sm font-medium text-text">{label}</p>
	<Skeleton shape="control" />
</div>
