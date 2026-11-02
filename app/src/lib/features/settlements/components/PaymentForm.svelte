<script lang="ts">
	import Copy from '@lucide/svelte/icons/copy';
	import { untrack } from 'svelte';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Card from '#lib/components/ui/Card.svelte';
	import DateField from '#lib/components/ui/DateField.svelte';
	import ImageField from '#lib/components/ui/ImageField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import { pdfType } from '#lib/api/files.js';
	import type { PaymentMethod, Settlement } from '#lib/api/settlements.js';
	import { formatRupiah } from '#lib/i18n/format.js';
	import { settlements as id } from '#lib/i18n/settlements.js';
	import { PaymentComposer, acceptedMethods } from '../payment.svelte';

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
	composer.method = untrack(() => acceptedMethods(settlement)[0] ?? 'CASH');
	const methodOptions = $derived(
		acceptedMethods(settlement).map((value) => ({ value, label: text.methods[value] }))
	);
	const payee = $derived(settlement.payee);
	const proofTypes = `image/jpeg,image/png,image/webp,${pdfType}`;

	async function copyAccount(number: string) {
		try {
			await navigator.clipboard.writeText(number);
			toaster.show({ tone: 'success', message: text.accountCopied });
		} catch {
			toaster.show({ tone: 'danger', message: text.copyFailed });
		}
	}

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
	<form onsubmit={handleSubmit} novalidate class="flex flex-col gap-5">
		<p class="flex flex-wrap items-baseline justify-between gap-2 text-base text-text">
			{text.amountToPay}
			<span class="font-mono text-lg font-bold tabular-nums">
				{formatRupiah(settlement.split.consignor_share)}
			</span>
		</p>
		<Select
			label={text.method}
			options={methodOptions}
			value={composer.method}
			error={composer.errors.method}
			onchange={(chosen: string) => (composer.method = chosen as PaymentMethod)}
		/>
		<section
			aria-label={text.payee}
			class="flex flex-col gap-4 rounded-control border border-border bg-accent-surface p-4"
		>
			<h3 class="text-base font-semibold text-text">{text.payee}</h3>
			{#if !payee}
				<p class="text-sm text-text-muted">{text.payeeEmpty}</p>
			{:else if composer.method === 'TRANSFER'}
				{#if payee.bank_account_number}
					<dl class="grid gap-4 text-sm md:grid-cols-3">
						<div class="flex flex-col gap-1">
							<dt class="text-text-muted">{text.bankName}</dt>
							<dd class="font-semibold text-text">{payee.bank_name}</dd>
						</div>
						<div class="flex flex-col gap-1">
							<dt class="text-text-muted">{text.bankAccountNumber}</dt>
							<dd class="flex flex-wrap items-center gap-2">
								<span class="font-mono text-base font-bold text-text tabular-nums"
									>{payee.bank_account_number}</span
								>
								<button
									type="button"
									aria-label={text.copyAccount}
									onclick={() => void copyAccount(payee.bank_account_number)}
									class="inline-flex size-11 items-center justify-center rounded-control text-accent hover:bg-surface"
									><Copy class="size-5" aria-hidden="true" /></button
								>
							</dd>
						</div>
						<div class="flex flex-col gap-1">
							<dt class="text-text-muted">{text.bankAccountHolder}</dt>
							<dd class="font-semibold text-text">{payee.bank_account_holder}</dd>
						</div>
					</dl>
				{/if}
				<p class="text-sm text-text">{text.transferGuide}</p>
			{:else if composer.method === 'QRIS'}
				{#if payee.payment_qr_file_id}
					<img
						src="/api/v1/files/{payee.payment_qr_file_id}"
						alt={text.qrAlt}
						width="192"
						height="192"
						class="h-auto w-full max-w-(--qr-code-size) rounded-control border border-border bg-surface object-contain"
					/>
				{/if}
				<p class="text-sm text-text">{text.qrisGuide}</p>
			{:else}
				<p class="text-sm text-text">{text.cashGuide}</p>
			{/if}
			{#if payee?.payment_info}
				<p class="text-sm whitespace-pre-line text-text">
					<span class="font-semibold">{text.payeeNote}:</span>
					{payee.payment_info}
				</p>
			{/if}
		</section>
		<DateField
			label={text.paidOnLabel}
			value={composer.paidOn}
			max={composer.today}
			error={composer.errors.paid_on}
			onchange={(paidOn: string) => (composer.paidOn = paidOn)}
		/>
		<ImageField
			label={text.proof}
			chooseLabel={text.chooseProof}
			imageUrl={composer.proof?.documentName ? undefined : composer.proof?.url}
			documentName={composer.proof?.documentName}
			imageAlt={text.proofAlt}
			accept={proofTypes}
			hint={text.proofHint}
			error={composer.errors.proof_file_id}
			busy={composer.isUploading}
			onselect={(file: File) => void composer.chooseProof(file)}
		/>
		{#if composer.errors.amount !== undefined || composer.failure !== ''}
			<Alert title={text.paymentFailed}>{composer.errors.amount ?? composer.failure}</Alert>
		{/if}
		<Button type="submit" busy={composer.busy}>{text.save}</Button>
	</form>
</Card>
