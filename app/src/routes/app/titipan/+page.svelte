<script lang="ts">
	import Plus from '@lucide/svelte/icons/plus';
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import Package from '@lucide/svelte/icons/package';
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
		<PageHeading icon={Package}>{id.drops.historyTitle}</PageHeading>
		{#if isStore}
			<a href="/app/titipan/baru" class="gap-2 font-semibold text-accent"
				><Plus class="size-5" aria-hidden="true" />{id.drops.recordDrop}</a
			>
		{/if}
	</div>
	<DropHistoryList recordHref={isStore ? '/app/titipan/baru' : undefined} />
</section>
