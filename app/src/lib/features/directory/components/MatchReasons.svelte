<script lang="ts">
	import type { MatchReason } from '#lib/api/directory.js';
	import { reasonText } from '../reasons';

	interface Props {
		reasons: MatchReason[];
		isApproximate?: boolean;
	}

	let { reasons, isApproximate = false }: Props = $props();
</script>

{#if reasons.length > 0}
	<ul class="flex flex-wrap gap-2 text-sm text-text">
		{#each reasons as reason (reason.code)}
			<li class="rounded-pill bg-surface-hover px-2 py-1">{reasonText(reason, isApproximate)}</li>
		{/each}
	</ul>
{/if}
