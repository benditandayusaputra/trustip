<script lang="ts">
	import CheckCheck from '@lucide/svelte/icons/check-check';
	import ArrowRight from '@lucide/svelte/icons/arrow-right';
	import { home as id } from '#lib/i18n/home.js';

	const text = id.home.example;
</script>

<aside class="soft-panel flex flex-col gap-5 p-6 shadow-card md:p-8" aria-label={text.title}>
	<div class="flex flex-col gap-1">
		<p class="section-label">{text.title}</p>
		<h3 class="text-xl font-semibold text-text">{text.partners}</h3>
		<p class="text-xs text-text-muted">{text.pricing}</p>
	</div>
	<div class="flex flex-wrap justify-between gap-4 border-b border-border pb-5">
		<div>
			<p class="font-semibold text-text">{text.product}</p>
			<p class="text-sm text-text-muted">{text.units}</p>
		</div>
		<p class="text-lg font-bold text-text tabular-nums">{text.gross}</p>
	</div>
	<dl class="grid grid-cols-2 gap-4 border-b border-border pb-5">
		<dt class="text-text-muted">{text.storeShareLabel}</dt>
		<dd class="text-right font-bold text-text tabular-nums">{text.storeShare}</dd>
	</dl>
	<div>
		<p class="text-sm text-text-muted">{text.mustPayLabel}</p>
		<p class="text-kpi font-bold text-text tabular-nums">{text.mustPay}</p>
	</div>
	<p
		class="flex items-start gap-2 rounded-control border border-border bg-accent-surface p-4 text-sm font-medium text-accent"
	>
		<CheckCheck class="size-5 shrink-0" aria-hidden="true" />{text.sameRecord}
	</p>
	<a
		href="/demo"
		class="inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-accent hover:underline"
	>
		{text.action}<ArrowRight class="size-4" aria-hidden="true" />
	</a>
</aside>
