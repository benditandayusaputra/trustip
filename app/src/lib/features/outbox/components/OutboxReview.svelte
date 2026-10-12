<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import ResponsiveTable from '#lib/components/ui/ResponsiveTable.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { offline as id } from '#lib/i18n/offline.js';
	import { findOwnEntry, outboxScheduler } from '#lib/offline/device.js';
	import { fixHrefOf } from '../review-rows';
	import { EntryReview } from '../review.svelte';

	interface Props {
		entryId: string;
		ondiscarded: () => void;
	}

	let { entryId, ondiscarded }: Props = $props();

	const text = id.outbox;
	const review = new EntryReview({
		client: apiClient,
		find: findOwnEntry,
		discard: (chosen) => outboxScheduler().discard(chosen)
	});
	const countColumns = [
		{ key: 'product' as const, label: text.product },
		{ key: 'seen' as const, label: text.seenStock, numeric: true },
		{ key: 'latest' as const, label: text.latestStock, numeric: true },
		{ key: 'counted' as const, label: text.countedQty, numeric: true },
		{ key: 'returned' as const, label: text.returnedQty, numeric: true }
	];
	const dropColumns = [
		{ key: 'product' as const, label: text.product },
		{ key: 'qty' as const, label: text.qty, numeric: true },
		{ key: 'expiry' as const, label: text.expiry },
		{ key: 'agreed' as const, label: text.agreed }
	];
	let confirming = $state(false);

	$effect(() => {
		void review.load(entryId);
	});

	async function discard() {
		confirming = false;
		await review.discard();
		ondiscarded();
	}
</script>

{#if review.status === 'loading'}
	<ContentSkeleton variant="detail" />
{:else if review.status === 'missing'}
	<Alert title={text.reviewMissing}>
		<a href="/app/antrean" class="font-semibold text-accent">{text.backToOutbox}</a>
	</Alert>
{:else if review.status === 'failed' || !review.entry}
	<Alert title={text.reviewFailed}>{review.failure}</Alert>
{:else}
	{@const entry = review.entry}
	<p class="text-base font-semibold text-text">{text.kinds[entry.kind]} · {entry.partnerName}</p>
	<p class="text-sm text-text-muted">{text.recordedAt(formatDateTime(entry.recordedAt))}</p>
	<Alert title={text.lastError(entry.problem?.message ?? '')}>{text.reviewIntro}</Alert>
	{#if review.countRows.length > 0}
		<ResponsiveTable
			caption={text.yourInput}
			columns={countColumns}
			rows={review.countRows}
			rowKey="productId"
		/>
	{/if}
	{#if review.dropRows.length > 0}
		<p class="text-sm text-text">{review.versionNote}</p>
		<ResponsiveTable
			caption={text.yourInput}
			columns={dropColumns}
			rows={review.dropRows}
			rowKey="productId"
		/>
	{/if}
	<div class="flex flex-wrap items-center gap-2">
		<a
			href={fixHrefOf(entry)}
			class="inline-flex min-h-12 items-center rounded-control bg-primary px-5 font-semibold text-on-accent"
		>
			{text.fix}
		</a>
		<Button variant="secondary" onclick={() => (confirming = true)}>{text.discard}</Button>
	</div>
{/if}

<Dialog bind:open={confirming} title={text.discardTitle}>
	<p class="text-base text-text">{text.discardBody}</p>
	{#snippet actions()}
		<Button variant="secondary" onclick={() => (confirming = false)}>{text.cancel}</Button>
		<Button variant="danger" onclick={() => void discard()}>{text.discard}</Button>
	{/snippet}
</Dialog>
