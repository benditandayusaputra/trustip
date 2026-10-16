<script lang="ts">
	import Megaphone from '@lucide/svelte/icons/megaphone';
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { NeedAction } from '#lib/api/needs.js';
	import { formatBasisPoints, formatDate, formatRupiah } from '#lib/i18n/format.js';
	import { needs as needsText } from '#lib/i18n/needs.js';
	import { proposals as proposalsText } from '#lib/i18n/proposals.js';
	import { NeedRecord } from '../state.svelte';
	import { needTone } from '../tone';

	const id = { ...needsText, ...proposalsText };

	interface Props {
		needId: string;
	}

	let { needId }: Props = $props();

	const text = id.needs;
	const record = new NeedRecord(apiClient);
	const toaster = getToaster();
	const done: Record<NeedAction, string> = {
		extend: text.extended,
		fulfill: text.fulfilled,
		close: text.closed
	};
	let confirming = $state<'fulfill' | 'close' | undefined>();
	let open = $derived(confirming !== undefined);

	$effect(() => {
		void record.load(needId);
	});

	async function apply(action: NeedAction) {
		const failure = await record.change(action);
		confirming = undefined;
		toaster.show(
			failure === undefined
				? { tone: 'success', message: done[action] }
				: { tone: 'danger', message: failure }
		);
	}
</script>

{#if record.status === 'loading'}
	<ContentSkeleton variant="detail" />
{:else if record.status === 'failed' || !record.detail}
	<Alert title={text.failed}>{record.failure}</Alert>
{:else}
	{@const need = record.detail}
	<div class="flex flex-wrap items-center gap-2">
		<PageHeading icon={Megaphone} backHref="/app/kebutuhan">{need.title}</PageHeading>
		<Badge tone={needTone(need.status)} label={text.statuses[need.status]} />
	</div>
	<p class="text-base text-text-muted">
		{need.category.name} · {text.expiresAt(formatDate(need.expires_at))}
	</p>
	{#if need.description}
		<p class="text-base text-text">{need.description}</p>
	{/if}
	<ul class="flex flex-wrap gap-2 text-sm text-text-muted">
		{#if need.est_qty_per_week !== null}
			<li>{text.perWeek(need.est_qty_per_week)}</li>
		{/if}
		{#if need.max_retail_price !== null}
			<li>{text.priceUpTo(formatRupiah(need.max_retail_price))}</li>
		{/if}
		{#if need.preferred_model}
			<li>{text.models[need.preferred_model]}</li>
		{/if}
		{#if need.offered_commission_bps !== null}
			<li>{formatBasisPoints(need.offered_commission_bps)}</li>
		{/if}
	</ul>
	<div class="flex flex-wrap gap-2">
		{#if need.status === 'OPEN' || need.status === 'EXPIRED'}
			<Button variant="secondary" busy={record.busy} onclick={() => void apply('extend')}>
				{text.extend}
			</Button>
		{/if}
		{#if need.status === 'OPEN'}
			<Button onclick={() => (confirming = 'fulfill')}>{text.fulfill}</Button>
			<Button variant="danger" onclick={() => (confirming = 'close')}>{text.close}</Button>
		{/if}
	</div>
	<section aria-labelledby="need-responses" class="flex flex-col gap-2">
		<h2 id="need-responses" class="text-lg font-semibold text-text">{text.responsesTitle}</h2>
		{#each need.responses as response (response.proposal_id)}
			<a
				href="/app/ajuan/{response.proposal_id}"
				class="surface-card flex items-center justify-between gap-4 p-5 hover:border-accent hover:bg-surface-hover"
			>
				<span class="flex flex-col">
					<span class="font-semibold text-text">{response.consignor_name}</span>
					<span class="font-mono text-sm text-text-muted">{response.number}</span>
				</span>
				<Badge tone="neutral" label={id.proposals.statuses[response.status]} />
			</a>
		{:else}
			<p class="text-base text-text-muted">{text.noResponses}</p>
		{/each}
	</section>
	{#if confirming}
		{@const chosen = confirming}
		<Dialog
			bind:open={() => open, (value: boolean) => (confirming = value ? confirming : undefined)}
			title={chosen === 'close' ? text.closeTitle : text.fulfillTitle}
		>
			<p class="text-base text-text">{chosen === 'close' ? text.closeBody : text.fulfillBody}</p>
			{#snippet actions()}
				<Button variant="secondary" onclick={() => (confirming = undefined)}>{text.cancel}</Button>
				<Button
					variant={chosen === 'close' ? 'danger' : 'primary'}
					busy={record.busy}
					onclick={() => void apply(chosen)}
				>
					{chosen === 'close' ? text.close : text.fulfill}
				</Button>
			{/snippet}
		</Dialog>
	{/if}
{/if}
