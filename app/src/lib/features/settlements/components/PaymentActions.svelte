<script lang="ts">
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { id } from '#lib/i18n/id.js';
	import { canMarkPaid, canReview, type Viewer } from '../access';
	import type { PaymentChoice, SettlementRecord } from '../record.svelte';

	interface Props {
		record: SettlementRecord;
		viewer: Viewer;
		reminderUrl?: string | undefined;
	}

	let { record, viewer, reminderUrl }: Props = $props();

	const text = id.settlements;
	const maxReasonLength = 200;
	const toaster = getToaster();
	const titles: Record<PaymentChoice, string> = {
		confirm: text.confirmTitle,
		reject: text.rejectTitle,
		markPaid: text.markPaidTitle
	};
	const bodies: Record<PaymentChoice, string> = {
		confirm: text.confirmBody,
		reject: text.rejectBody,
		markPaid: text.markPaidBody
	};
	const labels: Record<PaymentChoice, string> = {
		confirm: text.confirm,
		reject: text.reject,
		markPaid: text.markPaid
	};
	const done: Record<PaymentChoice, string> = {
		confirm: text.confirmed,
		reject: text.rejected,
		markPaid: text.markedPaid
	};
	let choice = $state<PaymentChoice | undefined>();
	let reason = $state('');
	let reasonError = $state<string | undefined>();
	let open = $derived(choice !== undefined);

	async function decide(chosen: PaymentChoice) {
		if (chosen === 'reject' && reason.trim() === '') {
			reasonError = text.reasonRequired;
			return;
		}
		const failure = await record.decide(chosen, reason.trim());
		choice = undefined;
		reason = '';
		reasonError = undefined;
		toaster.show(
			failure === undefined
				? { tone: 'success', message: done[chosen] }
				: { tone: 'danger', message: failure }
		);
	}
</script>

{#if record.settlement}
	{@const settlement = record.settlement}
	<div class="flex flex-wrap gap-2">
		{#if canReview(settlement, viewer)}
			<Button onclick={() => (choice = 'confirm')}>{text.confirm}</Button>
			<Button variant="danger" onclick={() => (choice = 'reject')}>{text.reject}</Button>
		{/if}
		{#if canMarkPaid(settlement, viewer)}
			<Button variant="secondary" onclick={() => (choice = 'markPaid')}>{text.markPaid}</Button>
		{/if}
		{#if reminderUrl}
			<a
				href={reminderUrl}
				target="_blank"
				rel="noopener noreferrer"
				class="inline-flex min-h-12 items-center rounded-control border border-border-strong px-5 font-semibold text-text"
			>
				{text.remind}
				<span class="sr-only">{id.auth.opensInNewTab}</span>
			</a>
		{/if}
	</div>
{/if}

{#if choice}
	{@const chosen = choice}
	<Dialog
		bind:open={() => open, (value: boolean) => (choice = value ? choice : undefined)}
		title={titles[chosen]}
	>
		<p class="text-base text-text">{bodies[chosen]}</p>
		{#if chosen === 'reject'}
			<TextField
				label={text.reason}
				bind:value={reason}
				maxlength={maxReasonLength}
				error={reasonError}
			/>
		{/if}
		{#snippet actions()}
			<Button variant="secondary" onclick={() => (choice = undefined)}>{text.keep}</Button>
			<Button
				variant={chosen === 'reject' ? 'danger' : 'primary'}
				busy={record.busy}
				onclick={() => void decide(chosen)}
			>
				{labels[chosen]}
			</Button>
		{/snippet}
	</Dialog>
{/if}
