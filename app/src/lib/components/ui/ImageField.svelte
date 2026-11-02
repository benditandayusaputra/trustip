<script lang="ts">
	import FileText from '@lucide/svelte/icons/file-text';
	import ImageUp from '@lucide/svelte/icons/image-up';
	import LoaderCircle from '@lucide/svelte/icons/loader-circle';
	import { common as id } from '#lib/i18n/common.js';

	interface Props {
		label: string;
		chooseLabel: string;
		imageUrl: string | undefined;
		imageAlt: string;
		documentName?: string | undefined;
		accept?: string;
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
		documentName,
		accept = 'image/jpeg,image/png,image/webp',
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
	{#if documentName}
		<p
			class="inline-flex w-fit max-w-full items-center gap-2 rounded-control border border-border bg-surface px-4 py-2 text-sm font-medium text-text"
		>
			<FileText class="size-5 shrink-0 text-accent" aria-hidden="true" /><span class="truncate"
				>{documentName}</span
			>
		</p>
	{:else if imageUrl}
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
			{accept}
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
