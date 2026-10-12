<script lang="ts">
	import { tick } from 'svelte';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Proposal } from '#lib/api/proposals.js';
	import { agreements as agreementsText } from '#lib/i18n/agreements.js';
	import { proposals as proposalsText } from '#lib/i18n/proposals.js';
	import { AmendmentComposer } from '../amendment.svelte';
	import AmendmentRows from './AmendmentRows.svelte';
	import GeneralTermsFields from './GeneralTermsFields.svelte';

	const id = { ...agreementsText, ...proposalsText };

	interface Props {
		partnershipId: string;
		prefillProductId: string;
		onsent: (proposal: Proposal) => void;
	}

	let { partnershipId, prefillProductId, onsent }: Props = $props();

	const text = id.proposals;
	const composer = new AmendmentComposer(apiClient);
	const toaster = getToaster();
	let chosen = $state('');
	let formElement: HTMLFormElement | undefined = $state();
	const addOptions = $derived([
		{ value: '', label: text.chooseProduct },
		...composer.addable.map((product) => ({ value: product.product_id, label: product.name }))
	]);

	$effect(() => {
		void composer.load(partnershipId, prefillProductId);
	});

	function addChosen() {
		const product = composer.addable.find((candidate) => candidate.product_id === chosen);
		if (product) {
			composer.add(product);
			chosen = '';
		}
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const outcome = await composer.send(partnershipId);
		if (outcome.kind === 'sent') {
			toaster.show({ tone: 'success', message: text.amendmentSent });
			onsent(outcome.proposal);
		} else if (outcome.kind === 'failed') {
			toaster.show({ tone: 'danger', message: outcome.message });
		} else {
			await tick();
			formElement?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
		}
	}
</script>

{#if composer.status === 'loading'}
	<ContentSkeleton variant="form" />
{:else if composer.status === 'failed'}
	<Alert title={id.agreement.loadFailed}>{composer.failure}</Alert>
{:else}
	<form bind:this={formElement} onsubmit={handleSubmit} novalidate class="flex flex-col gap-6">
		<fieldset class="flex flex-col gap-4">
			<legend class="mb-2 text-lg font-semibold text-text">{text.currentProducts}</legend>
			<AmendmentRows
				rows={composer.rows}
				errors={composer.errors}
				onremove={(productId: string) => {
					composer.remove(productId);
				}}
			/>
			{#if composer.errors.items}
				<p class="text-sm text-danger">{composer.errors.items}</p>
			{/if}
		</fieldset>
		<fieldset class="flex flex-col gap-2">
			<legend class="mb-2 text-lg font-semibold text-text">{text.addProducts}</legend>
			{#if composer.addable.length === 0}
				<p class="text-sm text-text-muted">{text.nothingToAdd}</p>
			{:else}
				<Select label={text.chooseProduct} options={addOptions} bind:value={chosen} />
				<Button variant="secondary" disabled={chosen === ''} onclick={addChosen}>
					{text.addProduct}
				</Button>
			{/if}
		</fieldset>
		<GeneralTermsFields values={composer.values} errors={composer.errors} />
		<TextField
			label={text.message}
			hint={text.messageHint}
			bind:value={composer.values.message}
			error={composer.errors.message}
		/>
		<Button type="submit" busy={composer.busy}>{text.sendAmendment}</Button>
	</form>
{/if}
