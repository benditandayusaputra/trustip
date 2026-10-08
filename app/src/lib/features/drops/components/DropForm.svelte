<script lang="ts">
	import { tick } from 'svelte';
	import PackageOpen from '@lucide/svelte/icons/package-open';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import DateField from '#lib/components/ui/DateField.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import { id } from '#lib/i18n/id.js';
	import { noteCompletedAction } from '#lib/offline/install.js';
	import { findOwnEntry, outboxScheduler } from '#lib/offline/device.js';
	import { DropComposer } from '../composer.svelte';
	import { previewOf, type DropLine } from '../schema';
	import DropLineRow from './DropLineRow.svelte';
	import DropPreview from './DropPreview.svelte';

	interface Props {
		partnershipId: string;
		onsaved: (dropId: string) => void;
		onqueued: () => void;
		revisionOf?: string;
	}

	let { partnershipId, onsaved, onqueued, revisionOf = '' }: Props = $props();

	const text = id.drops;
	const toaster = getToaster();
	const composer = new DropComposer({
		client: apiClient,
		outbox: {
			submit: (draft) => outboxScheduler().submit(draft),
			resubmit: (entryId, draft) => outboxScheduler().resubmit(entryId, draft),
			find: findOwnEntry
		},
		now: () => new Date()
	});
	const preview = $derived(previewOf(composer.lines));
	let formElement: HTMLFormElement | undefined = $state();

	$effect(() => {
		void (revisionOf === ''
			? composer.load(partnershipId)
			: composer.loadRevision(partnershipId, revisionOf));
	});

	async function repeatLast() {
		const number = await composer.repeatLast().catch((error: unknown) => {
			toaster.show({ tone: 'danger', message: error instanceof Error ? error.message : '' });
			return null;
		});
		if (number === undefined) {
			toaster.show({ tone: 'info', message: text.nothingToRepeat });
		} else if (number !== null) {
			toaster.show({ tone: 'info', message: text.repeated(number) });
		}
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const outcome = await composer.save();
		if (outcome.kind === 'saved' || outcome.kind === 'queued') {
			noteCompletedAction();
		}
		if (outcome.kind === 'saved') {
			toaster.show({ tone: 'success', message: text.saved(outcome.document.number) });
			onsaved(outcome.document.id);
		} else if (outcome.kind === 'queued') {
			toaster.show({ tone: 'info', message: id.offline.queued });
			onqueued();
		} else if (outcome.kind === 'failed') {
			toaster.show({ tone: 'danger', message: outcome.message });
		} else {
			await tick();
			formElement?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
		}
	}
</script>

{#if composer.status === 'loading'}
	<Skeleton />
{:else if composer.status === 'failed'}
	<Alert title={text.loadFailed}>{composer.failure}</Alert>
{:else if composer.lines.length === 0}
	<EmptyState icon={PackageOpen} message={text.noProducts}>
		{#snippet action()}
			<a href="/app/mitra/{partnershipId}" class="font-semibold text-accent">{id.nav.partners}</a>
		{/snippet}
	</EmptyState>
{:else}
	<form bind:this={formElement} onsubmit={handleSubmit} novalidate class="flex flex-col gap-6">
		{#if composer.isRevising}
			<p role="status" class="text-sm text-warning">{id.outbox.revising}</p>
		{/if}
		<p class="text-base text-text">{text.consignor}: {composer.consignorName}</p>
		<DateField
			label={text.dropDate}
			value={composer.dropDate}
			min={composer.earliestDate}
			max={composer.today}
			hint={text.dropDateHint}
			error={composer.errors.dropDate}
			onchange={(next: string) => {
				composer.changeDate(next);
			}}
		/>
		<Button variant="secondary" onclick={() => void repeatLast()}>{text.repeatLast}</Button>
		{#each composer.lines as line (line.productId)}
			<DropLineRow
				{line}
				suggestion={composer.suggestionFor(line.productId)}
				dropDate={composer.dropDate}
				errors={composer.errors}
				onchange={(changes: Partial<DropLine>) => {
					composer.updateLine(line.productId, changes);
				}}
			/>
		{/each}
		{#if composer.errors.items}
			<p class="text-sm text-danger" role="alert">{composer.errors.items}</p>
		{/if}
		<DropPreview {preview} />
		<Button type="submit" busy={composer.busy}>{text.save}</Button>
	</form>
{/if}
