<script lang="ts">
	import Bell from '@lucide/svelte/icons/bell';
	import { apiClient } from '#lib/api/client.js';
	import { id } from '#lib/i18n/id.js';
	import { UnreadCounter } from '../inbox.svelte';

	const counter = new UnreadCounter(apiClient);
	const badgeLimit = 99;

	$effect(() => {
		void counter.refresh();
	});
</script>

<a
	href="/app/pemberitahuan"
	aria-label={id.notifications.bell(counter.count)}
	class="relative inline-flex size-11 items-center justify-center rounded-control text-text"
>
	<Bell class="size-6" aria-hidden="true" />
	{#if counter.count > 0}
		<span
			aria-hidden="true"
			class="absolute top-1 right-1 min-w-5 rounded-pill bg-accent px-1 text-center text-xs font-semibold text-on-accent"
		>
			{counter.count > badgeLimit ? `${String(badgeLimit)}+` : counter.count}
		</span>
	{/if}
</a>
