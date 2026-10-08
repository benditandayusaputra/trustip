<script lang="ts">
	import SettlementList from '#lib/features/settlements/components/SettlementList.svelte';
	import { id } from '#lib/i18n/id.js';
</script>

<svelte:head>
	<title>{id.settlements.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.settlements.title}</h1>
	<SettlementList />
</section>
