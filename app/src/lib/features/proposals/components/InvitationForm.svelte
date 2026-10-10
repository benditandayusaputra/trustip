<script lang="ts">
	import { tick } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Checkbox from '#lib/components/ui/Checkbox.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Proposal } from '#lib/api/proposals.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { proposals as id } from '#lib/i18n/proposals.js';
	import { InvitationComposer } from '../composer.svelte';
	import GeneralTermsFields from './GeneralTermsFields.svelte';
	import ItemTermsRow from './ItemTermsRow.svelte';

	interface Props {
		isConsignor: boolean;
		targetId?: string;
		needId?: string;
		preselectedProductId?: string;
		submitLabel?: string;
		sentMessage?: string;
		onsent: (proposal: Proposal) => void;
	}

	let {
		isConsignor,
		targetId = '',
		needId = '',
		preselectedProductId = '',
		submitLabel = id.proposals.send,
		sentMessage = id.proposals.sent,
		onsent
	}: Props = $props();

	const text = id.proposals;
	const composer = new InvitationComposer(
		apiClient,
		() => targetId,
		() => needId
	);
	const listsProducts = $derived(isConsignor || targetId !== '');
	const toaster = getToaster();
	let formElement: HTMLFormElement | undefined = $state();

	$effect(() => {
		if (isConsignor) {
			void composer.loadCatalog();
		} else if (targetId !== '') {
			void composer.loadTargetCatalog(preselectedProductId);
		}
	});

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const outcome = await composer.send();
		if (outcome.kind === 'sent') {
			toaster.show({ tone: 'success', message: sentMessage });
			onsent(outcome.proposal);
		} else if (outcome.kind === 'failed') {
			toaster.show({ tone: 'danger', message: outcome.message });
		} else {
			await tick();
			formElement?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
		}
	}
</script>

<form bind:this={formElement} onsubmit={handleSubmit} novalidate class="flex flex-col gap-6">
	{#if listsProducts}
		<fieldset class="flex flex-col gap-2">
			<legend class="mb-2 text-lg font-semibold text-text">{text.products}</legend>
			<p class="text-sm text-text-muted">{text.productsHint}</p>
			{#each composer.catalog as entry (entry.id)}
				<Checkbox
					checked={composer.isChosen(entry.id)}
					onchange={(isChosen: boolean) => {
						composer.choose(entry, isChosen);
					}}
				>
					{entry.name} · {formatRupiah(entry.retailPrice)}
				</Checkbox>
			{:else}
				<p class="text-sm text-warning">{text.noProducts}</p>
			{/each}
		</fieldset>
		{#each composer.values.items as item, index (item.productId)}
			<ItemTermsRow {item} {index} errors={composer.errors} />
		{/each}
	{:else}
		<p class="text-sm text-text-muted">{text.storeIntro}</p>
	{/if}
	<GeneralTermsFields values={composer.values} errors={composer.errors} />
	<TextField
		label={text.message}
		hint={text.messageHint}
		bind:value={composer.values.message}
		error={composer.errors.message}
	/>
	<Button type="submit" busy={composer.busy}>{submitLabel}</Button>
</form>
