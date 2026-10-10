<script lang="ts">
	import CloudUpload from '@lucide/svelte/icons/cloud-upload';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { offline as id } from '#lib/i18n/offline.js';
	import { watchOutbox } from '#lib/offline/announcer.js';
	import { outboxScheduler, ownEntries } from '#lib/offline/device.js';
	import type { OutboxEntry } from '#lib/offline/outbox.js';
	import { OutboxQueue } from '../queue.svelte';
	import OutboxEntryCard from './OutboxEntryCard.svelte';

	const text = id.outbox;
	const toaster = getToaster();
	const scheduler = outboxScheduler();
	const queue = new OutboxQueue({
		list: ownEntries,
		send: () => scheduler.run(),
		discard: (entryId) => scheduler.discard(entryId)
	});
	let discarding = $state<OutboxEntry>();
	let confirming = $state(false);
	const oldCount = $derived(queue.oldCount(new Date()));

	$effect(() => {
		void queue.load();
		return watchOutbox(() => void queue.load());
	});

	async function sendNow() {
		toaster.show({ tone: 'info', message: await queue.sendNow() });
	}

	function askDiscard(entry: OutboxEntry) {
		discarding = entry;
		confirming = true;
	}

	async function discard(entry: OutboxEntry) {
		confirming = false;
		await queue.discard(entry.id);
		toaster.show({ tone: 'info', message: text.discarded });
	}
</script>

<div class="flex flex-wrap items-center gap-4">
	<Button busy={queue.sending} onclick={() => void sendNow()}>{text.sendNow}</Button>
	<p class="text-sm text-text-muted">{text.iosNote}</p>
</div>
{#if oldCount > 0}
	<Alert title={text.oldWarning(oldCount)}>{text.iosNote}</Alert>
{/if}
{#if queue.status === 'loading'}
	<Skeleton />
{:else if queue.status === 'failed'}
	<Alert title={text.loadFailed}>{queue.failure}</Alert>
{:else if queue.entries.length === 0}
	<EmptyState icon={CloudUpload} message={text.empty}>
		{#snippet action()}
			<a href="/app" class="font-semibold text-accent">{text.backToApp}</a>
		{/snippet}
	</EmptyState>
{:else}
	<ul class="flex flex-col gap-2">
		{#each queue.entries as entry (entry.id)}
			<OutboxEntryCard {entry} ondiscard={askDiscard} />
		{/each}
	</ul>
{/if}

<Dialog bind:open={confirming} title={text.discardTitle}>
	<p class="text-base text-text">{text.discardBody}</p>
	{#snippet actions()}
		<Button variant="secondary" onclick={() => (confirming = false)}>{text.cancel}</Button>
		<Button
			variant="danger"
			onclick={() => {
				if (discarding) {
					void discard(discarding);
				}
			}}>{text.discard}</Button
		>
	{/snippet}
</Dialog>
