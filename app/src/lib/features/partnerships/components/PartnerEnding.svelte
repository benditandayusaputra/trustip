<script lang="ts">
	import Button from '#lib/components/ui/Button.svelte';
	import Card from '#lib/components/ui/Card.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Partnership } from '#lib/api/partnerships.js';
	import { id } from '#lib/i18n/id.js';
	import { EndingRecord, blockerHref } from '../ending.svelte';

	interface Props {
		partnership: Partnership;
		onchanged: (partnership: Partnership) => void;
	}

	let { partnership, onchanged }: Props = $props();

	const text = id.partners;
	const maxReasonLength = 500;
	const toaster = getToaster();
	const record = new EndingRecord(apiClient);
	let open = $state(false);
	let reason = $state('');

	$effect(() => {
		void record.load(partnership.id);
	});

	async function end() {
		const outcome = await record.end(partnership.id, reason.trim());
		open = false;
		if (outcome.kind === 'failed') {
			toaster.show({ tone: 'danger', message: outcome.message });
			return;
		}
		toaster.show({ tone: 'success', message: text.ended });
		onchanged(outcome.partnership);
	}
</script>

<Card title={text.end}>
	<p class="text-sm text-text-muted">{text.endIntro}</p>
	{#if record.status === 'loading'}
		<Skeleton />
	{:else if record.status === 'failed'}
		<p class="text-sm text-text">{text.blockersFailed}</p>
	{:else if record.blockers.length === 0}
		<p class="text-sm text-text">{text.noBlockers}</p>
	{:else}
		<h3 class="text-base font-semibold text-text">{text.blockers}</h3>
		<ul class="flex list-disc flex-col gap-1 pl-5 text-sm text-text">
			{#each record.blockers as blocker (`${blocker.kind}-${blocker.document_id}`)}
				{@const href = blockerHref(blocker)}
				<li>
					{#if href}
						<a {href} class="font-semibold text-accent">{blocker.description}</a>
					{:else}
						{blocker.description}
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
	<Button variant="danger" onclick={() => (open = true)}>{text.end}</Button>
</Card>

<Dialog bind:open title={text.endTitle}>
	<p class="text-base text-text">{text.endBody}</p>
	<TextField label={text.endReason} bind:value={reason} maxlength={maxReasonLength} />
	{#snippet actions()}
		<Button variant="secondary" onclick={() => (open = false)}>{text.keep}</Button>
		<Button variant="danger" busy={record.busy} onclick={() => void end()}>{text.end}</Button>
	{/snippet}
</Dialog>
