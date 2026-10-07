<script lang="ts">
	import ImageUp from '@lucide/svelte/icons/image-up';
	import LoaderCircle from '@lucide/svelte/icons/loader-circle';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		label: string;
		chooseLabel: string;
		imageUrl: string | undefined;
		imageAlt: string;
		hint?: string | undefined;
		error?: string | undefined;
		busy?: boolean;
		onselect: (file: File) => void;
	}

	let {
		label,
		chooseLabel,
		imageUrl,
		imageAlt,
		hint,
		error,
		busy = false,
		onselect
	}: Props = $props();

	const inputId = $props.id();
	const previewSize = 48;

	let input: HTMLInputElement;

	function pick() {
		const file = input.files?.[0];
		input.value = '';
		if (file) {
			onselect(file);
		}
	}
</script>

<div class="flex flex-col gap-2">
	<p class="text-sm font-medium text-text">{label}</p>
	{#if imageUrl}
		<img
			src={imageUrl}
			alt={imageAlt}
			width={previewSize}
			height={previewSize}
			decoding="async"
			class="size-12 rounded-control border border-border object-cover"
		/>
	{/if}
	<label
		for={inputId}
		class="inline-flex min-h-12 w-fit cursor-pointer items-center gap-2 rounded-control border border-border-strong bg-bg px-5 text-base font-semibold text-text focus-within:outline-2 focus-within:outline-accent"
	>
		{#if busy}
			<LoaderCircle class="size-5 animate-spin" aria-hidden="true" />
			{id.form.saving}
		{:else}
			<ImageUp class="size-5" aria-hidden="true" />
			{chooseLabel}
		{/if}
		<input
			bind:this={input}
			id={inputId}
			type="file"
			accept="image/jpeg,image/png,image/webp"
			disabled={busy}
			aria-invalid={error ? true : undefined}
			aria-describedby={[hint ? `${inputId}-hint` : '', error ? `${inputId}-error` : '']
				.filter(Boolean)
				.join(' ') || undefined}
			onchange={pick}
			class="sr-only"
		/>
	</label>
	{#if hint}
		<p id="{inputId}-hint" class="text-xs text-text-muted">{hint}</p>
	{/if}
	{#if error}
		<p id="{inputId}-error" class="text-sm text-danger">{error}</p>
	{/if}
</div>
