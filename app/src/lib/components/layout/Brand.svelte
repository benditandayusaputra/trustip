<script lang="ts">
	import { common as text } from '#lib/i18n/common.js';

	let {
		href = '/',
		compact = false,
		subtitle
	}: { href?: string; compact?: boolean; subtitle?: string } = $props();
</script>

<a
	{href}
	aria-label={text.brand}
	class="brand inline-flex min-h-11 items-center gap-2 text-xl font-semibold tracking-(--letter-spacing-heading)"
>
	<img src="/icons/icon-192.png" alt="" width="32" height="32" class="size-8 shrink-0" />
	{#if subtitle}
		<span class="flex flex-col leading-tight"
			>{text.brand}<span class="brand-subtitle text-xs font-medium tracking-normal">{subtitle}</span
			></span
		>
	{:else}
		<span class={compact ? 'hidden compact:inline' : ''}>{text.brand}</span>
	{/if}
</a>
