<script lang="ts">
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import CountHistoryList from '#lib/features/counts/components/CountHistoryList.svelte';
	import { counts as id } from '#lib/i18n/counts.js';

	const session = getAccountSession();
	const isStore = $derived(session.account?.membership?.organization_type === 'STORE');
</script>

<svelte:head>
	<title>{id.counts.historyTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<h1 class="text-xl font-semibold text-text">{id.counts.historyTitle}</h1>
		{#if isStore}
			<a href="/app/hitung/baru" class="font-semibold text-accent">{id.counts.count}</a>
		{/if}
	</div>
	<CountHistoryList countHref={isStore ? '/app/hitung/baru' : undefined} />
</section>
