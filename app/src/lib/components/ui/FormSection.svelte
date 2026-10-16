<script lang="ts">
	import type { Snippet } from 'svelte';
	let { title, description, children }: { title: string; description?: string; children: Snippet } =
		$props();
</script>

<section class="form-section flex flex-col gap-5">
	<header class="flex flex-col gap-2">
		<h2 class="text-lg font-semibold text-text">{title}</h2>
		{#if description}<p class="text-sm text-text-muted">{description}</p>{/if}
	</header>
	{@render children()}
</section>
