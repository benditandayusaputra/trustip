<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Card from '#lib/components/ui/Card.svelte';
	import DateField from '#lib/components/ui/DateField.svelte';
	import ImageField from '#lib/components/ui/ImageField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { PaymentMethod, Settlement } from '#lib/api/settlements.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { PaymentComposer, needsProof } from '../payment.svelte';

	interface Props {
		settlement: Settlement;
		onsaved: (settlement: Settlement) => void;
	}

	let { settlement, onsaved }: Props = $props();

	const text = id.settlements;
	const toaster = getToaster();
	const composer = new PaymentComposer({
		client: apiClient,
		newKey: () => crypto.randomUUID(),
		now: () => new Date()
	});
	const methodOptions = Object.entries(text.methods).map(([value, label]) => ({ value, label }));
	const qrFileId = $derived(settlement.payee?.payment_qr_file_id ?? null);

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const saved = await composer.submit(settlement);
		if (saved) {
			toaster.show({ tone: 'success', message: text.saved });
			onsaved(saved);
		}
	}
</script>

<Card title={text.recordPayment}>
	<div class="flex flex-col gap-2">
		<h3 class="text-base font-semibold text-text">{text.payee}</h3>
		{#if settlement.payee && (settlement.payee.payment_info || qrFileId)}
			<p class="text-base whitespace-pre-line text-text">{settlement.payee.payment_info}</p>
			{#if qrFileId}
				<img
					src="/api/v1/files/{qrFileId}"
					alt={text.qrAlt}
					width="192"
					height="192"
					class="h-auto w-full max-w-form rounded-control border border-border object-contain"
				/>
			{/if}
		{:else}
			<p class="text-sm text-text-muted">{text.payeeEmpty}</p>
		{/if}
	</div>
	<form onsubmit={handleSubmit} novalidate class="flex flex-col gap-5">
		<p class="font-mono text-base text-text tabular-nums">
			{text.amountToPay(formatRupiah(settlement.split.consignor_share))}
		</p>
		<Select
			label={text.method}
			options={methodOptions}
			value={composer.method}
			error={composer.errors.method}
			onchange={(chosen: string) => (composer.method = chosen as PaymentMethod)}
		/>
		<DateField
			label={text.paidOnLabel}
			value={composer.paidOn}
			max={composer.today}
			error={composer.errors.paid_on}
			onchange={(paidOn: string) => (composer.paidOn = paidOn)}
		/>
		{#if needsProof(composer.method)}
			<ImageField
				label={text.proof}
				chooseLabel={text.chooseProof}
				imageUrl={composer.proof?.url}
				imageAlt={text.proofAlt}
				hint={text.proofHint}
				error={composer.errors.proof_file_id}
				busy={composer.isUploading}
				onselect={(file: File) => void composer.chooseProof(file)}
			/>
		{/if}
		{#if composer.errors.amount !== undefined || composer.failure !== ''}
			<Alert title={text.paymentFailed}>{composer.errors.amount ?? composer.failure}</Alert>
		{/if}
		<Button type="submit" busy={composer.busy}>{text.save}</Button>
	</form>
</Card>
