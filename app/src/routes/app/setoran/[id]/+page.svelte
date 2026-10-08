<script lang="ts">
	import { page } from '$app/state';
	import SettlementView from '#lib/features/settlements/components/SettlementView.svelte';
	import { id } from '#lib/i18n/id.js';
</script>

<svelte:head>
	<title>{id.settlements.detailTitle}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.settlements.detailTitle}</h1>
	{#key page.params.id}
		<SettlementView settlementId={page.params.id ?? ''} />
	{/key}
</section>
