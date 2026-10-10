<script lang="ts">
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { Tone } from '#lib/components/ui/tone.js';
	import type { Payment } from '#lib/api/settlements.js';
	import { formatBusinessDate, formatRupiah } from '#lib/i18n/format.js';
	import { common as commonText } from '#lib/i18n/common.js';
	import { settlements as settlementsText } from '#lib/i18n/settlements.js';

	const id = { ...commonText, ...settlementsText };

	interface Props {
		payments: Payment[];
	}

	let { payments }: Props = $props();

	const text = id.settlements;
	const tones: Record<Payment['status'], Tone> = {
		PENDING_CONFIRMATION: 'pending',
		CONFIRMED: 'success',
		REJECTED: 'danger'
	};
</script>

{#if payments.length > 0}
	<h2 class="text-lg font-semibold text-text">{text.payments}</h2>
	<ul class="flex flex-col gap-2">
		{#each payments as payment (payment.id)}
			<li class="flex flex-col gap-1 rounded-control border border-border bg-surface p-4">
				<div class="flex flex-wrap items-center justify-between gap-2">
					<span class="font-mono text-base text-text tabular-nums">
						{formatRupiah(payment.amount)}
					</span>
					<Badge tone={tones[payment.status]} label={text.paymentStatuses[payment.status]} />
				</div>
				<span class="text-sm text-text-muted">
					{text.paymentLine(
						text.methods[payment.method],
						formatBusinessDate(payment.paid_on),
						payment.recorded_by_name
					)}
				</span>
				{#if payment.rejection_reason}
					<span class="text-sm text-text">{text.rejectedBecause(payment.rejection_reason)}</span>
				{/if}
				{#if payment.proof_file_id}
					<a
						href="/api/v1/files/{payment.proof_file_id}"
						target="_blank"
						rel="noopener noreferrer"
						class="text-sm font-semibold text-accent"
					>
						{text.openProof}
						<span class="sr-only">{id.auth.opensInNewTab}</span>
					</a>
				{/if}
			</li>
		{/each}
	</ul>
{/if}
