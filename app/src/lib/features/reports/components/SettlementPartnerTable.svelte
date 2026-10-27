<script lang="ts">
	import type { SettlementReport } from '#lib/api/reports.js';
	import { formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { reports as id } from '#lib/i18n/reports.js';
	import { settlementKinds as kinds } from '../rows';

	interface Props {
		rows: SettlementReport['rows'];
		partnerCount: number;
	}

	let { rows, partnerCount }: Props = $props();

	const text = id.reports;
</script>

<section class="surface-card overflow-hidden" aria-labelledby="ringkasan-mitra">
	<header class="flex flex-wrap items-start justify-between gap-4 p-5">
		<div class="flex flex-col gap-1">
			<h2 id="ringkasan-mitra" class="text-lg font-semibold text-text">
				{text.partnerSummary}
			</h2>
			<p class="text-xs text-text-muted">{text.partnerSummaryHint}</p>
		</div>
		<span class="rounded-pill bg-surface-hover px-4 py-1 text-xs text-text"
			>{text.partnerCount(partnerCount)}</span
		>
	</header>
	{#if partnerCount === 0}
		<p class="px-5 pb-5 text-sm text-text-muted">{text.settlementsEmpty}</p>
	{:else if rows.length === 0}
		<p class="px-5 pb-5 text-sm text-text-muted">{text.noPartnerMatch}</p>
	{:else}
		<div class="hidden overflow-x-auto md:block">
			<table class="w-full border-collapse text-sm text-text">
				<caption class="sr-only">{text.partnerSummary}</caption>
				<thead>
					<tr
						class="border-y border-border bg-surface-hover text-xs tracking-(--letter-spacing-label) text-text-muted uppercase"
					>
						<th scope="col" class="px-5 py-4 text-left font-semibold">{text.partner}</th>
						{#each kinds as kind (kind.key)}
							<th scope="col" class="px-5 py-4 text-right font-semibold">{kind.label}</th>
						{/each}
					</tr>
				</thead>
				<tbody>
					{#each rows as row (row.partner_id)}
						<tr class="border-b border-border last:border-b-0 hover:bg-surface-hover">
							<th scope="row" class="px-5 py-4 text-left font-semibold">{row.name}</th>
							{#each kinds as kind (kind.key)}
								<td class="px-5 py-4 text-right"
									><span
										class="tabular-nums {row[kind.key].count > 0 ? kind.tone : 'text-text-muted'}"
										>{formatInteger(row[kind.key].count)} · {formatRupiah(
											row[kind.key].amount
										)}</span
									></td
								>
							{/each}
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
		<ul class="flex flex-col border-t border-border md:hidden" aria-label={text.partnerSummary}>
			{#each rows as row (row.partner_id)}
				<li class="flex flex-col gap-2 border-b border-border p-5 last:border-b-0">
					<p class="font-semibold text-text">{row.name}</p>
					<dl class="grid grid-cols-2 gap-2 text-sm">
						{#each kinds as kind (kind.key)}
							<dt class="text-text-muted">{kind.label}</dt>
							<dd class="text-right">
								<span class="tabular-nums {row[kind.key].count > 0 ? kind.tone : 'text-text-muted'}"
									>{formatInteger(row[kind.key].count)} · {formatRupiah(row[kind.key].amount)}</span
								>
							</dd>
						{/each}
					</dl>
				</li>
			{/each}
		</ul>
	{/if}
</section>
