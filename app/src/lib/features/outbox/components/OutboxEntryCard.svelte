<script lang="ts">
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import type { OutboxEntry } from '#lib/offline/outbox.js';
	import { statusToneOf } from '../queue.svelte';

	interface Props {
		entry: OutboxEntry;
		ondiscard: (entry: OutboxEntry) => void;
	}

	let { entry, ondiscard }: Props = $props();

	const text = id.outbox;
	const documentPaths = { DROP: '/app/titipan', COUNT: '/app/hitung' };
</script>

<li class="flex flex-col gap-2 rounded-card border border-border bg-surface p-4">
	<div class="flex flex-wrap items-center justify-between gap-2">
		<p class="text-base font-semibold text-text">
			{text.kinds[entry.kind]} · {entry.partnerName}
		</p>
		<Badge tone={statusToneOf(entry)} label={text.statuses[entry.status]} />
	</div>
	<p class="text-sm text-text-muted">{text.recordedAt(formatDateTime(entry.recordedAt))}</p>
	{#if entry.status === 'HELD'}
		<p class="text-sm text-text-muted">{text.heldHint}</p>
	{:else if entry.problem && entry.status !== 'SENT'}
		<p class="text-sm text-danger">{text.lastError(entry.problem.message)}</p>
	{/if}
	<div class="flex flex-wrap items-center gap-2">
		{#if entry.document}
			<a href="{documentPaths[entry.kind]}/{entry.document.id}" class="font-semibold text-accent"
				>{text.openDocument(entry.document.number)}</a
			>
		{:else}
			{#if entry.status === 'NEEDS_REVIEW'}
				<a href="/app/antrean/{entry.id}" class="font-semibold text-accent">{text.review}</a>
			{/if}
			<Button
				variant="secondary"
				onclick={() => {
					ondiscard(entry);
				}}>{text.discard}</Button
			>
		{/if}
	</div>
</li>
