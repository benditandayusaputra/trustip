<script lang="ts">
	import type { Snippet } from 'svelte';

	let { filters, children }: { filters: Snippet; children: Snippet } = $props();
</script>

<div class="grid items-start gap-6 xl:grid-cols-(--directory-columns)">
	<div class="trade-panel">{@render filters()}</div>
	<div class="flex min-w-0 flex-col gap-4">{@render children()}</div>
</div>
