<script lang="ts">
	import { tick } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Checkbox from '#lib/components/ui/Checkbox.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Proposal } from '#lib/api/proposals.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { InvitationComposer } from '../composer.svelte';
	import GeneralTermsFields from './GeneralTermsFields.svelte';
	import ItemTermsRow from './ItemTermsRow.svelte';

	interface Props {
		isConsignor: boolean;
		onsent: (proposal: Proposal) => void;
	}

	let { isConsignor, onsent }: Props = $props();

	const text = id.proposals;
	const composer = new InvitationComposer(apiClient);
	const toaster = getToaster();
	let formElement: HTMLFormElement | undefined = $state();

	$effect(() => {
		if (isConsignor) {
			void composer.loadCatalog();
		}
	});

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const outcome = await composer.send();
		if (outcome.kind === 'sent') {
			toaster.show({ tone: 'success', message: text.sent });
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
	{#if isConsignor}
		<fieldset class="flex flex-col gap-2">
			<legend class="mb-2 text-lg font-semibold text-text">{text.products}</legend>
			<p class="text-sm text-text-muted">{text.productsHint}</p>
			{#each composer.catalog as product (product.id)}
				<Checkbox
					checked={composer.isChosen(product.id)}
					onchange={(isChosen: boolean) => {
						composer.choose(product, isChosen);
					}}
				>
					{product.name} · {formatRupiah(product.suggested_retail_price)}
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
	<Button type="submit" busy={composer.busy}>{text.send}</Button>
</form>
