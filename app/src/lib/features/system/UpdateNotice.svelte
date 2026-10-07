<script lang="ts">
	import { updated } from '$app/state';
	import Button from '#lib/components/ui/Button.svelte';
	import { id } from '#lib/i18n/id.js';
</script>

{#if updated.current}
	<div
		role="status"
		class="fixed inset-x-0 top-0 z-(--z-toast) flex items-center justify-center gap-4 border-b border-border bg-accent-surface px-4 py-2"
	>
		<p class="text-base text-text">{id.system.updateAvailable}</p>
		<Button
			variant="secondary"
			onclick={() => {
				location.reload();
			}}>{id.system.reload}</Button
		>
	</div>
{/if}
