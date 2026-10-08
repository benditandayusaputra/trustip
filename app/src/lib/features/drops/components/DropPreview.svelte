<script lang="ts">
	import { formatInteger, formatRupiah } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import type { DropPreview } from '../schema';

	interface Props {
		preview: DropPreview;
	}

	let { preview }: Props = $props();

	const text = id.drops;
	const figures = $derived([
		{ label: text.units, value: formatInteger(preview.units) },
		{ label: text.gross, value: formatRupiah(preview.split.gross) },
		{ label: text.storeShare, value: formatRupiah(preview.split.storeShare) },
		{ label: text.consignorShare, value: formatRupiah(preview.split.consignorShare) }
	]);
</script>

<section
	aria-live="polite"
	class="flex flex-col gap-2 rounded-card border border-border bg-surface p-4"
>
	<h2 class="text-base font-semibold text-text">{text.preview}</h2>
	<dl class="grid grid-cols-2 gap-2 text-sm">
		{#each figures as figure (figure.label)}
			<dt class="text-text-muted">{figure.label}</dt>
			<dd class="text-right font-mono font-medium text-text tabular-nums">{figure.value}</dd>
		{/each}
	</dl>
</section>
