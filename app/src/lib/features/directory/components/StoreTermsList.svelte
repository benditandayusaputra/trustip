<script lang="ts">
	import type { PublicStore } from '#lib/api/public.js';
	import { formatBasisPoints } from '#lib/i18n/format.js';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { publicProfile as publicProfileText } from '#lib/i18n/public-profile.js';

	const id = { ...directoryText, ...publicProfileText };

	interface Props {
		terms: PublicStore['terms'];
		otherRequirements: string;
	}

	let { terms, otherRequirements }: Props = $props();

	const text = id.publicProfile;
	const lines = $derived(
		[
			terms.accepts_fixed_deposit ? text.fixedDeposit : '',
			terms.accepts_commission ? text.commission : '',
			terms.commission_min_bps !== null && terms.commission_max_bps !== null
				? id.directory.commissionRange(
						formatBasisPoints(terms.commission_min_bps),
						formatBasisPoints(terms.commission_max_bps)
					)
				: '',
			terms.requires_distribution_permit ? text.requiresPermit : '',
			terms.requires_halal ? text.requiresHalal : '',
			terms.requires_sample ? text.requiresSample : ''
		].filter((line) => line !== '')
	);
</script>

{#if lines.length === 0 && otherRequirements === ''}
	<p class="text-base text-text-muted">{text.noTerms}</p>
{:else}
	<ul class="flex list-disc flex-col gap-1 pl-5 text-base text-text">
		{#each lines as line (line)}
			<li>{line}</li>
		{/each}
	</ul>
	{#if otherRequirements}
		<p class="text-base text-text">
			<span class="text-text-muted">{text.otherRequirements}:</span>
			{otherRequirements}
		</p>
	{/if}
{/if}
