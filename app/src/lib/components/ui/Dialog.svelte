<script lang="ts">
	import type { Snippet } from 'svelte';
	import CloseIcon from '@lucide/svelte/icons/x';
	import { common as id } from '#lib/i18n/common.js';

	interface Props {
		open?: boolean;
		title: string;
		placement?: 'center' | 'bottom';
		children: Snippet;
		actions?: Snippet | undefined;
	}

	let { open = $bindable(false), title, placement = 'center', children, actions }: Props = $props();

	const titleId = $props.id();
	let dialog: HTMLDialogElement | undefined = $state();

	const placements = {
		center: 'm-auto w-full max-w-form rounded-card',
		bottom: 'mx-auto mt-auto mb-0 w-full max-w-form rounded-t-card'
	};

	$effect(() => {
		if (open && !dialog?.open) {
			dialog?.showModal();
		}
		if (!open && dialog?.open) {
			dialog.close();
		}
	});
</script>

<dialog
	bind:this={dialog}
	aria-labelledby={titleId}
	onclose={() => (open = false)}
	class="border border-border bg-surface p-0 text-text backdrop:bg-scrim {placements[placement]}"
>
	<div class="flex flex-col gap-4 p-5">
		<header class="flex items-center justify-between gap-4">
			<h2 id={titleId} class="text-lg">{title}</h2>
			<button
				type="button"
				class="inline-flex size-11 items-center justify-center text-text-muted"
				aria-label={id.feedback.close}
				onclick={() => (open = false)}
			>
				<CloseIcon class="size-5" aria-hidden="true" />
			</button>
		</header>
		<div>{@render children()}</div>
		{#if actions}
			<footer class="flex flex-col-reverse gap-2 md:flex-row md:justify-end">
				{@render actions()}
			</footer>
		{/if}
	</div>
</dialog>
