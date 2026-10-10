<script lang="ts">
	import type { Tally } from '#lib/domain/count.js';
	import { formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { counts as id } from '#lib/i18n/counts.js';

	interface Props {
		tally: Tally;
	}

	let { tally }: Props = $props();

	const text = id.counts;
	const figures = $derived([
		{ label: text.soldUnits, value: formatInteger(tally.soldQty) },
		{ label: text.returnedUnits, value: formatInteger(tally.returnedQty) },
		{ label: text.gross, value: formatRupiah(tally.split.gross) },
		{ label: text.storeShare, value: formatRupiah(tally.split.storeShare) },
		{ label: text.consignorShare, value: formatRupiah(tally.split.consignorShare) }
	]);
</script>

<section
	aria-live="polite"
	class="flex flex-col gap-4 rounded-card border border-border bg-accent-surface p-5"
>
	<h2 class="text-base font-semibold text-text">{text.preview}</h2>
	<dl class="grid grid-cols-2 gap-4 text-sm">
		{#each figures as figure (figure.label)}
			<dt class="text-text-muted">{figure.label}</dt>
			<dd class="text-right font-mono text-base font-semibold text-text tabular-nums">
				{figure.value}
			</dd>
		{/each}
	</dl>
</section>
