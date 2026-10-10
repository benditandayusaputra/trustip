<script lang="ts">
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import DropHistoryList from '#lib/features/drops/components/DropHistoryList.svelte';
	import { drops as id } from '#lib/i18n/drops.js';

	const session = getAccountSession();
	const isStore = $derived(session.account?.membership?.organization_type === 'STORE');
</script>

<svelte:head>
	<title>{id.drops.historyTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<h1 class="text-xl font-semibold text-text">{id.drops.historyTitle}</h1>
		{#if isStore}
			<a href="/app/titipan/baru" class="font-semibold text-accent">{id.drops.recordDrop}</a>
		{/if}
	</div>
	<DropHistoryList recordHref={isStore ? '/app/titipan/baru' : undefined} />
</section>
