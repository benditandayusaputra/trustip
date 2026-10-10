<script lang="ts">
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { proposals as id } from '#lib/i18n/proposals.js';
	import type { ProposalRecord } from '../record.svelte';

	interface Props {
		record: ProposalRecord;
		canDecide: boolean;
		canWithdraw: boolean;
		oncounter: () => void;
	}

	let { record, canDecide, canWithdraw, oncounter }: Props = $props();

	type Choice = 'accept' | 'decline' | 'withdraw';

	const text = id.proposals;
	const toaster = getToaster();
	const done: Record<Choice, string> = {
		accept: text.accepted,
		decline: text.declined,
		withdraw: text.withdrawn
	};
	let choice = $state<Choice | undefined>();
	let reason = $state('');
	let open = $derived(choice !== undefined);

	async function confirm(chosen: Choice) {
		const actions = {
			accept: () => record.accept(),
			decline: () => record.decline(reason),
			withdraw: () => record.withdraw()
		};
		const failure = await actions[chosen]();
		if (failure === undefined) {
			choice = undefined;
			toaster.show({ tone: 'success', message: done[chosen] });
		} else if (!record.reasonError) {
			toaster.show({ tone: 'danger', message: failure });
		}
	}
</script>

<div class="flex flex-wrap gap-2">
	{#if canDecide}
		<Button onclick={() => (choice = 'accept')}>{text.accept}</Button>
		<Button variant="secondary" onclick={oncounter}>{text.counter}</Button>
		<Button variant="danger" onclick={() => (choice = 'decline')}>{text.decline}</Button>
	{/if}
	{#if canWithdraw}
		<Button variant="secondary" onclick={() => (choice = 'withdraw')}>{text.withdraw}</Button>
	{/if}
</div>

{#if choice}
	{@const chosen = choice}
	<Dialog
		bind:open={() => open, (value: boolean) => (choice = value ? choice : undefined)}
		title={{ accept: text.acceptTitle, decline: text.declineTitle, withdraw: text.withdrawTitle }[
			chosen
		]}
	>
		{#if chosen === 'accept'}
			<p class="text-base text-text">{text.acceptBody}</p>
		{:else if chosen === 'withdraw'}
			<p class="text-base text-text">{text.withdrawBody}</p>
		{:else}
			<TextField
				label={text.reason}
				hint={text.reasonHint}
				bind:value={reason}
				error={record.reasonError}
			/>
		{/if}
		{#snippet actions()}
			<Button variant="secondary" onclick={() => (choice = undefined)}>{text.cancel}</Button>
			<Button
				variant={chosen === 'accept' ? 'primary' : 'danger'}
				busy={record.busy}
				onclick={() => void confirm(chosen)}
			>
				{{ accept: text.accept, decline: text.decline, withdraw: text.withdraw }[chosen]}
			</Button>
		{/snippet}
	</Dialog>
{/if}
