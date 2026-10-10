<script lang="ts">
	import { common as text } from '#lib/i18n/common.js';

	let { href = '/' }: { href?: string } = $props();
</script>

<a
	{href}
	class="inline-flex min-h-11 items-center gap-2 text-xl font-semibold tracking-(--letter-spacing-heading) text-text"
>
	<img src="/icons/icon.svg" alt="" width="32" height="32" class="size-8 shrink-0" />
	{text.brand}
</a>
