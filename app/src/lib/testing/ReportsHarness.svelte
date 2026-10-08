<script lang="ts">
	import ReportsPage from '#lib/features/reports/components/ReportsPage.svelte';

	interface Props {
		organizationType?: 'STORE' | 'CONSIGNOR';
	}

	let { organizationType = 'STORE' }: Props = $props();
</script>

<ReportsPage {organizationType} />
