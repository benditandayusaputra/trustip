<script lang="ts">
	import SalesReportView from '#lib/features/reports/components/SalesReportView.svelte';

	interface Props {
		organizationType?: 'STORE' | 'CONSIGNOR';
	}

	let { organizationType = 'STORE' }: Props = $props();
</script>

<SalesReportView {organizationType} />
