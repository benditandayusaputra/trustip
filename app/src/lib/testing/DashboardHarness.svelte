<script lang="ts">
	import { createRawSnippet } from 'svelte';
	import type { SalesNarrowing } from '#lib/api/reports.js';
	import ConsignorDashboard from '#lib/features/dashboard/components/ConsignorDashboard.svelte';

	interface Props {
		isOwner?: boolean;
	}

	let { isOwner = true }: Props = $props();

	const range = { from: '2026-10-22', to: '2026-11-20' };
	const chart = createRawSnippet((narrowing: () => SalesNarrowing) => ({
		render: () => `<p>chart ${narrowing().partnerId ?? narrowing().productId ?? 'all'}</p>`
	}));
</script>

<ConsignorDashboard {isOwner} {range} {chart} />
