<script lang="ts">
	import { getAccountSession } from '#lib/features/account/session.svelte.js';
	import SalesReportView from '#lib/features/reports/components/SalesReportView.svelte';
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
	<h1 class="text-xl font-semibold text-text">{id.reports.salesTitle}</h1>
	{#key organizationType}
		<SalesReportView {organizationType} />
	{/key}
</section>
