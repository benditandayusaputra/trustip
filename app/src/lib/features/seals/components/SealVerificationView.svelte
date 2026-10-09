<script lang="ts">
	import ShieldAlert from '@lucide/svelte/icons/shield-alert';
	import ShieldCheck from '@lucide/svelte/icons/shield-check';
	import ShieldQuestion from '@lucide/svelte/icons/shield-question';
	import type { SealVerification } from '#lib/api/public.js';
	import { formatDateTime } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		code: string;
		verification: SealVerification | null;
	}

	let { code, verification }: Props = $props();

	const text = id.seal;
	const isValid = $derived(verification?.status === 'VALID');
</script>

<article class="mx-auto flex w-full max-w-form flex-col gap-6">
	<header class="flex flex-col gap-2">
		<p class="text-sm text-text-muted">{text.verifyTitle}</p>
		<h1 class="font-mono text-xl font-semibold text-text">{verification?.code ?? code}</h1>
	</header>
	{#if verification === null}
		<section class="flex flex-col gap-2 rounded-card border border-border bg-surface p-4">
			<p class="flex items-center gap-2 text-base font-semibold text-text">
				<ShieldQuestion class="size-6 text-text-muted" aria-hidden="true" />
				{text.notFound}
			</p>
			<p class="text-sm text-text-muted">{text.notFoundBody}</p>
		</section>
	{:else}
		<section
			class="flex flex-col gap-4 rounded-card border p-4 {isValid
				? 'border-success bg-success-surface'
				: 'border-danger bg-danger-surface'}"
		>
			<p
				class="flex items-center gap-2 text-base font-semibold {isValid
					? 'text-success'
					: 'text-danger'}"
			>
				{#if isValid}
					<ShieldCheck class="size-6" aria-hidden="true" />
				{:else}
					<ShieldAlert class="size-6" aria-hidden="true" />
				{/if}
				{isValid ? text.valid : text.invalid}
			</p>
			<p class="text-sm text-text">{isValid ? text.publicValidBody : text.publicInvalidBody}</p>
		</section>
		<dl class="grid gap-2 text-base">
			<dt class="text-text-muted">{text.document}</dt>
			<dd class="text-text">
				{text.documentTypes[verification.document_type] ?? verification.document_type}
				<span class="font-mono">{verification.document_number}</span>
			</dd>
			<dt class="text-text-muted">{text.event}</dt>
			<dd class="text-text">
				{text.eventNames[verification.event_type] ?? verification.event_type}
			</dd>
			<dt class="text-text-muted">{text.time}</dt>
			<dd class="text-text">{formatDateTime(verification.occurred_at)}</dd>
			<dt class="text-text-muted">{text.parties}</dt>
			<dd class="text-text">
				{text.partiesValue(verification.consignor_name, verification.store_name)}
			</dd>
		</dl>
	{/if}
	<p class="text-sm text-text-muted">{text.limits}</p>
</article>
