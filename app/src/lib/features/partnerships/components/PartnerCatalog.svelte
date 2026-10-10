<script lang="ts">
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { PartnerCatalogItem } from '#lib/api/partnerships.js';
	import { formatBasisPoints, formatRupiah } from '#lib/i18n/format.js';
	import { partnerships as partnershipsText } from '#lib/i18n/partnerships.js';
	import { proposals as proposalsText } from '#lib/i18n/proposals.js';

	const id = { ...partnershipsText, ...proposalsText };

	interface Props {
		items: PartnerCatalogItem[];
		proposeHref?: ((productId: string) => string) | undefined;
	}

	let { items, proposeHref }: Props = $props();

	const text = id.partners;
	const terms = id.proposals;
</script>

<section class="flex flex-col gap-2">
	<h2 class="text-lg font-semibold text-text">{text.catalog}</h2>
	<ul class="flex flex-col gap-2">
		{#each items as item (item.product_id)}
			<li class="flex flex-col gap-1 rounded-control border border-border bg-surface p-4">
				<span class="flex flex-wrap items-center justify-between gap-2">
					<span class="text-base font-semibold text-text">{item.name}</span>
					{#if !item.terms}
						<Badge tone="neutral" label={text.notAgreed} />
						{#if proposeHref}
							<a href={proposeHref(item.product_id)} class="text-sm font-semibold text-accent">
								{id.proposals.proposeToStore}
							</a>
						{/if}
					{:else if !item.is_active}
						<Badge tone="neutral" label={text.stoppedItem} />
					{/if}
				</span>
				<span class="text-sm text-text-muted">{item.sku} · {item.category_name} · {item.unit}</span>
				{#if item.terms}
					<span class="text-sm text-text">
						{item.terms.pricing_model === 'COMMISSION'
							? `${terms.models.COMMISSION} ${formatBasisPoints(item.terms.commission_bps)}`
							: terms.models.FIXED_DEPOSIT} · {terms.retailPrice}
						<span class="font-mono tabular-nums">{formatRupiah(item.terms.retail_price)}</span>
						· {terms.perUnitConsignor}
						<span class="font-mono tabular-nums">
							{formatRupiah(item.terms.per_unit.consignor_share)}
						</span>
					</span>
				{/if}
			</li>
		{/each}
	</ul>
</section>
