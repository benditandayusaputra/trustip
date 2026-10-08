<script lang="ts">
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import ReportsPage from '#lib/features/reports/components/ReportsPage.svelte';
	import { id } from '#lib/i18n/id.js';

	const session = getAccountSession();
	const organizationType = $derived(
		session.account?.membership?.organization_type === 'CONSIGNOR' ? 'CONSIGNOR' : 'STORE'
	);
</script>

<svelte:head>
	<title>{id.reports.title}</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-content flex-col gap-6">
	<h1 class="text-xl font-semibold text-text">{id.reports.title}</h1>
	{#key organizationType}
		<ReportsPage {organizationType} />
	{/key}
</section>
