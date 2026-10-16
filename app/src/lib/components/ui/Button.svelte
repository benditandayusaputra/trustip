<script lang="ts">
	import type { Snippet } from 'svelte';
	import LoaderCircle from '@lucide/svelte/icons/loader-circle';
	import { common as id } from '#lib/i18n/common.js';

	type Variant = 'primary' | 'secondary' | 'danger';

	interface Props {
		children: Snippet;
		variant?: Variant;
		type?: 'button' | 'submit';
		disabled?: boolean;
		busy?: boolean;
		onclick?: (event: MouseEvent) => void;
	}

	let {
		children,
		variant = 'primary',
		type = 'button',
		disabled = false,
		busy = false,
		onclick
	}: Props = $props();

	const variants: Record<Variant, string> = {
		primary: 'bg-primary text-on-accent border border-primary hover:bg-primary-hover',
		secondary: 'bg-surface text-text border border-border hover:bg-surface-hover',
		danger: 'bg-bg text-danger border border-danger'
	};
</script>

<button
	{type}
	{onclick}
	disabled={disabled || busy}
	aria-busy={busy}
	class="button-control inline-flex min-h-12 items-center justify-center gap-2 rounded-control px-5 text-base font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-60 {variants[
		variant
	]}"
>
	{#if busy}
		<LoaderCircle class="size-5 animate-spin" aria-hidden="true" />
		{id.form.saving}
	{:else}
		{@render children()}
	{/if}
</button>
