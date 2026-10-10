<script lang="ts">
	import CheckCheck from '@lucide/svelte/icons/check-check';
	import ArrowRight from '@lucide/svelte/icons/arrow-right';
	import { home as id } from '#lib/i18n/home.js';

	const text = id.home.example;
</script>

<aside class="trade-panel flex flex-col gap-5" aria-label={text.title}>
	<p class="section-label">{text.title}</p>
	<h2 class="text-lg font-semibold text-text">{text.partners}</h2>
	<div class="flex flex-wrap justify-between gap-4 border-b border-border pb-4">
		<div>
			<p class="font-semibold text-text">{text.product}</p>
			<p class="text-sm text-text-muted">{text.units}</p>
		</div>
		<p class="font-mono text-text tabular-nums">{text.gross}</p>
	</div>
	<dl class="grid grid-cols-2 gap-4">
		<dt class="text-sm text-text-muted">{text.storeShareLabel}</dt>
		<dd class="text-right font-mono text-text tabular-nums">{text.storeShare}</dd>
	</dl>
	<div>
		<p class="text-sm text-text-muted">{text.mustPayLabel}</p>
		<p class="mt-2 font-mono text-kpi font-bold text-text tabular-nums">{text.mustPay}</p>
	</div>
	<p class="flex items-start gap-2 rounded-control bg-accent-surface p-4 text-sm text-accent">
		<CheckCheck class="size-5 shrink-0" aria-hidden="true" />{text.sameRecord}
	</p>
	<a
		href="/masuk#coba-demo"
		class="inline-flex min-h-11 items-center gap-2 font-semibold text-accent"
	>
		{text.action}<ArrowRight class="size-4" aria-hidden="true" />
	</a>
	<p class="text-xs text-text-muted">{text.disclaimer}</p>
</aside>
