<script lang="ts">
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import PackageOpen from '@lucide/svelte/icons/package-open';
	import { dashboard as id } from '#lib/i18n/dashboard.js';
	import type { RankedRow } from '../ranking';
	interface Props {
		row: RankedRow;
		index: number;
	}
	let { row, index }: Props = $props();
</script>

<div class="flex min-w-0 flex-1 items-center gap-4">
	{#if row.product}
		<div
			class="flex size-12 shrink-0 items-center justify-center overflow-hidden rounded-control bg-surface-hover"
		>
			{#if row.product.photo_url}
				<img
					src={row.product.photo_url}
					alt=""
					width="48"
					height="48"
					loading="lazy"
					decoding="async"
					class="size-full object-cover"
				/>
			{:else}
				<PackageOpen class="size-6 text-text-muted" aria-hidden="true" />
			{/if}
		</div>
	{/if}
	<span class="flex min-w-0 flex-col gap-1">
		<span class="font-semibold wrap-anywhere">{index + 1}. {row.label}</span>
		<span class="font-mono tabular-nums md:hidden">{row.value}</span>
		{#if row.product}<span class="text-sm text-accent">{id.dashboard.productDetail}</span>{/if}
	</span>
</div>
<span class="flex shrink-0 items-center gap-4">
	<span class="hidden font-mono tabular-nums md:inline">{row.value}</span>
	{#if row.product}<ChevronRight class="size-5 text-text-muted" aria-hidden="true" />{/if}
</span>
