<script lang="ts">
	import { tick } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Proposal } from '#lib/api/proposals.js';
	import { id } from '#lib/i18n/id.js';
	import { CounterComposer } from '../counter.svelte';
	import GeneralTermsFields from './GeneralTermsFields.svelte';
	import ItemTermsRow from './ItemTermsRow.svelte';

	interface Props {
		proposal: Proposal;
		counterpartName: string;
		oncountered: (proposal: Proposal) => void;
		oncancel: () => void;
	}

	let { proposal, counterpartName, oncountered, oncancel }: Props = $props();

	const text = id.proposals;
	const composer = new CounterComposer(apiClient);
	const toaster = getToaster();
	let formElement: HTMLFormElement | undefined = $state();

	$effect.pre(() => {
		composer.start(proposal);
	});

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const outcome = await composer.send(proposal.id);
		if (outcome.kind === 'sent') {
			toaster.show({ tone: 'success', message: text.countered });
			oncountered(outcome.proposal);
		} else if (outcome.kind === 'failed') {
			toaster.show({ tone: 'danger', message: outcome.message });
		} else {
			await tick();
			formElement?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
		}
	}
</script>

<form
	bind:this={formElement}
	onsubmit={handleSubmit}
	novalidate
	aria-labelledby="counter-title"
	class="flex flex-col gap-6 rounded-card border border-accent p-4"
>
	<h2 id="counter-title" class="text-lg font-semibold text-text">{text.counterTitle}</h2>
	<p class="text-base text-text-muted">{text.counterIntro(counterpartName)}</p>
	{#each composer.values.items as item, index (item.productId)}
		{#if composer.actionOf(item.productId) === 'STOP'}
			<p class="text-base text-text">{item.name} · {text.modes.STOP}</p>
		{:else}
			<ItemTermsRow {item} {index} errors={composer.errors} />
		{/if}
		<Button
			variant="secondary"
			onclick={() => {
				composer.remove(item.productId);
			}}>{text.removeItem(item.name)}</Button
		>
	{/each}
	<GeneralTermsFields values={composer.values} errors={composer.errors} />
	<TextField
		label={text.message}
		hint={text.messageHint}
		bind:value={composer.values.message}
		error={composer.errors.message}
	/>
	<div class="flex flex-wrap gap-2">
		<Button type="submit" busy={composer.busy}>{text.counterSend}</Button>
		<Button variant="secondary" onclick={oncancel}>{text.counterCancel}</Button>
	</div>
</form>
