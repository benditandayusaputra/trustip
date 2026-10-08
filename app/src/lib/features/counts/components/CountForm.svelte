<script lang="ts">
	import { tick } from 'svelte';
	import Warehouse from '@lucide/svelte/icons/warehouse';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import DateField from '#lib/components/ui/DateField.svelte';
	import EmptyState from '#lib/components/ui/EmptyState.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { CountOccasion } from '#lib/api/counts.js';
	import { id } from '#lib/i18n/id.js';
	import { noteCompletedAction } from '#lib/offline/install.js';
	import { findOwnEntry, outboxScheduler, waitingEntries } from '#lib/offline/device.js';
	import { CountComposer } from '../composer.svelte';
	import { maxNoteLength, previewOf, type InputMode, type withChanges } from '../schema';
	import CountLineRow from './CountLineRow.svelte';
	import CountPreview from './CountPreview.svelte';

	interface Props {
		partnershipId: string;
		dropHref: string;
		onsaved: (countId: string) => void;
		onqueued: () => void;
		revisionOf?: string;
	}

	let { partnershipId, dropHref, onsaved, onqueued, revisionOf = '' }: Props = $props();

	const text = id.counts;
	const toaster = getToaster();
	const composer = new CountComposer({
		client: apiClient,
		outbox: {
			submit: (draft) => outboxScheduler().submit(draft),
			resubmit: (entryId, draft) => outboxScheduler().resubmit(entryId, draft),
			find: findOwnEntry,
			waiting: waitingEntries
		},
		now: () => new Date()
	});
	const preview = $derived(previewOf(composer.lines));
	const occasionOptions = Object.entries(text.occasions).map(([value, label]) => ({
		value,
		label
	}));
	const modeOptions = Object.entries(text.modes).map(([value, label]) => ({ value, label }));
	let formElement: HTMLFormElement | undefined = $state();

	$effect(() => {
		void (revisionOf === ''
			? composer.load(partnershipId)
			: composer.loadRevision(partnershipId, revisionOf));
	});

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
{:else if composer.status === 'failed' || !composer.sheet}
	<Alert title={text.loadFailed}>{composer.failure}</Alert>
{:else if !composer.sheet.is_active}
	<Alert title={text.notActive}>{composer.sheet.consignor.name}</Alert>
{:else if composer.lines.length === 0}
	<EmptyState icon={Warehouse} message={text.noStock}>
		{#snippet action()}
			<a href={dropHref} class="font-semibold text-accent">{text.newDrop}</a>
		{/snippet}
	</EmptyState>
{:else}
	<form bind:this={formElement} onsubmit={handleSubmit} novalidate class="flex flex-col gap-6">
		{#if composer.isRevising}
			<p role="status" class="text-sm text-warning">{id.outbox.revising}</p>
		{/if}
		<p class="text-base text-text">{text.consignor}: {composer.sheet.consignor.name}</p>
		<Select
			label={text.occasion}
			options={occasionOptions}
			value={composer.occasion}
			onchange={(occasion: string) => (composer.occasion = occasion as CountOccasion)}
		/>
		<DateField
			label={text.countedOn}
			bind:value={composer.countedOn}
			min={composer.earliestDate}
			max={composer.today}
			hint={text.countedOnHint}
			error={composer.errors.countedOn}
		/>
		<Select
			label={text.mode}
			options={modeOptions}
			value={composer.mode}
			onchange={(mode: string) => (composer.mode = mode as InputMode)}
		/>
		<div class="flex flex-wrap gap-2">
			<Button
				variant="secondary"
				onclick={() => {
					composer.markAllSold();
				}}>{text.allSold}</Button
			>
			<Button
				variant="secondary"
				onclick={() => {
					composer.returnAllRemaining();
				}}
			>
				{text.returnAll}
			</Button>
		</div>
		{#each composer.lines as line (line.productId)}
			<CountLineRow
				{line}
				mode={composer.mode}
				errors={composer.errors}
				onchange={(changes: Parameters<typeof withChanges>[1]) => {
					composer.updateLine(line.productId, changes);
				}}
			/>
		{/each}
		<TextField
			label={text.note}
			hint={text.noteHint}
			maxlength={maxNoteLength}
			bind:value={composer.note}
			error={composer.errors.note}
		/>
		<CountPreview tally={preview} />
		<Button type="submit" busy={composer.busy}>{text.save}</Button>
	</form>
{/if}
