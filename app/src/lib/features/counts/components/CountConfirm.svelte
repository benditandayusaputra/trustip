<script lang="ts">
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { counts as id } from '#lib/i18n/counts.js';
	import { noteCompletedAction } from '#lib/offline/install.js';
	import { canConfirm, type CountRecord } from '../record.svelte';

	interface Props {
		record: CountRecord;
		viewerOrganizationId: string;
	}

	let { record, viewerOrganizationId }: Props = $props();

	const text = id.counts;
	const toaster = getToaster();
	let open = $state(false);

	async function confirm() {
		const failure = await record.confirm();
		open = false;
		if (failure === undefined) {
			noteCompletedAction();
		}
		toaster.show(
			failure === undefined
				? { tone: 'success', message: text.confirmed }
				: { tone: 'danger', message: failure }
		);
	}
</script>

{#if record.count && canConfirm(record.count, viewerOrganizationId)}
	<Button onclick={() => (open = true)}>{text.confirm}</Button>
{/if}

{#if open}
	<Dialog bind:open title={text.confirmTitle}>
		<p class="text-base text-text">{text.confirmBody}</p>
		{#snippet actions()}
			<Button variant="secondary" onclick={() => (open = false)}>{text.keep}</Button>
			<Button busy={record.busy} onclick={() => void confirm()}>{text.confirm}</Button>
		{/snippet}
	</Dialog>
{/if}
