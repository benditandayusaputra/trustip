<script lang="ts">
	import CloudUpload from '@lucide/svelte/icons/cloud-upload';
	import { id } from '#lib/i18n/id.js';
	import { watchOutbox } from '#lib/offline/announcer.js';
	import { outboxScheduler, ownEntries } from '#lib/offline/device.js';
	import { OutboxQueue } from '../queue.svelte';

	const scheduler = outboxScheduler();
	const queue = new OutboxQueue({
		list: ownEntries,
		send: () => scheduler.run(),
		discard: (entryId) => scheduler.discard(entryId)
	});

	const hasOld = $derived(queue.oldCount(new Date()) > 0);

	$effect(() => {
		void queue.load();
		return watchOutbox(() => void queue.load());
	});
</script>

{#if queue.waitingCount > 0}
	<a
		href="/app/antrean"
		aria-label={id.outbox.badge(queue.waitingCount)}
		class="relative inline-flex size-11 items-center justify-center rounded-control {hasOld
			? 'text-danger'
			: 'text-warning'}"
	>
		<CloudUpload class="size-6" aria-hidden="true" />
		<span
			aria-hidden="true"
			class="absolute top-1 right-1 min-w-5 rounded-pill px-1 {hasOld
				? 'bg-danger'
				: 'bg-warning'} text-center text-xs font-semibold text-on-accent"
		>
			{queue.waitingCount}
		</span>
	</a>
{/if}
