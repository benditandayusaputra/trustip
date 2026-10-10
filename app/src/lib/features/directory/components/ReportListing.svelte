<script lang="ts">
	import Flag from '@lucide/svelte/icons/flag';
	import { onMount } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import { listingReport as id } from '#lib/i18n/listing-report.js';
	import { ListingReportForm, type ReportReason, type ReportTarget } from '../report.svelte';

	interface Props {
		target: ReportTarget;
		name: string;
		startOpen?: boolean;
	}

	let { target, name, startOpen = false }: Props = $props();

	const text = id.listingReport;
	const maxNoteLength = 500;
	const toaster = getToaster();
	const form = new ListingReportForm(apiClient);
	const reasonOptions = [
		{ value: '', label: text.chooseReason },
		...Object.entries(text.reasons).map(([value, label]) => ({ value, label }))
	];
	let open = $state(false);

	onMount(() => {
		open = startOpen;
	});

	async function send() {
		const outcome = await form.send(target);
		if (outcome.kind === 'sent') {
			open = false;
			toaster.show({ tone: 'success', message: text.sent });
		} else if (outcome.kind === 'failed') {
			toaster.show({ tone: 'danger', message: outcome.message });
		}
	}
</script>

<button
	type="button"
	class="inline-flex min-h-11 items-center gap-2 text-sm text-text-muted"
	onclick={() => (open = true)}
>
	<Flag class="size-4" aria-hidden="true" />
	{text.open(name)}
</button>

<Dialog bind:open title={text.title(name)}>
	<div class="flex flex-col gap-4">
		<p class="text-sm text-text-muted">{text.intro}</p>
		<Select
			label={text.reason}
			options={reasonOptions}
			bind:value={
				() => form.reason,
				(next: string) => {
					form.reason = next as ReportReason | '';
				}
			}
			required
			error={form.errors.reason}
		/>
		<TextField
			label={text.note}
			bind:value={form.note}
			maxlength={maxNoteLength}
			hint={text.noteHint}
			error={form.errors.note}
		/>
	</div>
	{#snippet actions()}
		<Button variant="secondary" onclick={() => (open = false)}>{text.cancel}</Button>
		<Button busy={form.busy} onclick={() => void send()}>{text.send}</Button>
	{/snippet}
</Dialog>
