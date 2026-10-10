<script lang="ts">
	import CloseIcon from '@lucide/svelte/icons/x';
	import { getToaster, type Toast } from './toast.svelte';
	import { edgeTone } from './tone';
	import { common as id } from '#lib/i18n/common.js';

	const toastLifetimeMs = 6000;
	const toaster = getToaster();

	function dismissLater(toast: Toast) {
		return () => {
			const timer = setTimeout(() => {
				toaster.dismiss(toast.id);
			}, toastLifetimeMs);
			return () => {
				clearTimeout(timer);
			};
		};
	}
</script>

<section
	aria-label={id.feedback.notifications}
	aria-live="polite"
	class="pointer-events-none fixed inset-x-0 bottom-(--toast-offset) z-(--z-toast) flex flex-col items-center gap-2 px-4 lg:bottom-8"
>
	{#each toaster.items as toast (toast.id)}
		<div
			{@attach dismissLater(toast)}
			class="pointer-events-auto flex w-full max-w-form items-center gap-4 rounded-control border border-l-4 border-border bg-surface p-4 {edgeTone[
				toast.tone
			]}"
		>
			<p class="flex-1 text-base">{toast.message}</p>
			{#if toast.action?.href}
				<a class="font-semibold text-accent" href={toast.action.href}>{toast.action.label}</a>
			{:else if toast.action?.onclick}
				<button type="button" class="font-semibold text-accent" onclick={toast.action.onclick}>
					{toast.action.label}
				</button>
			{/if}
			<button
				type="button"
				class="inline-flex size-11 items-center justify-center text-text-muted"
				aria-label={id.feedback.dismissToast}
				onclick={() => {
					toaster.dismiss(toast.id);
				}}
			>
				<CloseIcon class="size-5" aria-hidden="true" />
			</button>
		</div>
	{/each}
</section>
