<script lang="ts">
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { drops as id } from '#lib/i18n/drops.js';
	import { noteCompletedAction } from '#lib/offline/install.js';
	import { canCancel, canConfirm, type DropChoice, type DropRecord } from '../record.svelte';

	interface Props {
		record: DropRecord;
		viewerOrganizationId: string;
		now?: () => Date;
	}

	let { record, viewerOrganizationId, now = () => new Date() }: Props = $props();

	const text = id.drops;
	const toaster = getToaster();
	const titles: Record<DropChoice, string> = {
		confirm: text.confirmTitle,
		cancel: text.cancelTitle
	};
	const bodies: Record<DropChoice, string> = { confirm: text.confirmBody, cancel: text.cancelBody };
	const labels: Record<DropChoice, string> = { confirm: text.confirm, cancel: text.cancel };
	const done: Record<DropChoice, string> = { confirm: text.confirmed, cancel: text.cancelled };
	let choice = $state<DropChoice | undefined>();
	let open = $derived(choice !== undefined);

	async function decide(chosen: DropChoice) {
		const failure = await record.decide(chosen);
		choice = undefined;
		if (failure === undefined) {
			noteCompletedAction();
		}
		toaster.show(
			failure === undefined
				? { tone: 'success', message: done[chosen] }
				: { tone: 'danger', message: failure }
		);
	}
</script>

{#if record.drop}
	{@const drop = record.drop}
	<div class="flex flex-col gap-2">
		{#if canConfirm(drop, viewerOrganizationId)}
			<Button onclick={() => (choice = 'confirm')}>{text.confirm}</Button>
		{/if}
		{#if canCancel(drop, viewerOrganizationId, now())}
			<Button variant="danger" onclick={() => (choice = 'cancel')}>{text.cancel}</Button>
			<p class="text-xs text-text-muted">{text.cancelHint(formatDateTime(drop.cancel_deadline))}</p>
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
		{#snippet actions()}
			<Button variant="secondary" onclick={() => (choice = undefined)}>{text.keep}</Button>
			<Button
				variant={chosen === 'confirm' ? 'primary' : 'danger'}
				busy={record.busy}
				onclick={() => void decide(chosen)}
			>
				{labels[chosen]}
			</Button>
		{/snippet}
	</Dialog>
{/if}
