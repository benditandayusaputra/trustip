<script lang="ts">
	import type { SettlementNarrowing } from '#lib/api/settlements.js';
	import ReportsPage from '#lib/features/reports/components/ReportsPage.svelte';

	interface Props {
		organizationType?: 'STORE' | 'CONSIGNOR';
	}

	let { organizationType = 'STORE' }: Props = $props();
</script>

<ReportsPage {organizationType}>
	{#snippet settlements(narrowing: SettlementNarrowing)}<p>
			daftar setoran {narrowing.query}
		</p>{/snippet}
</ReportsPage>
