<script lang="ts">
	import type { PublicConsignor } from '#lib/api/public.js';
	import { publicProfile as id } from '#lib/i18n/public-profile.js';

	interface Props {
		legality: PublicConsignor['legality'];
	}

	let { legality }: Props = $props();

	const text = id.publicProfile;
	const entries = $derived(
		[
			{ label: text.businessId, value: legality.business_id_no },
			{ label: text.distributionPermit, value: legality.distribution_permit_no },
			{ label: text.halalCert, value: legality.halal_cert_no }
		].filter((entry) => entry.value !== null)
	);
</script>

{#if entries.length === 0}
	<p class="text-base text-text-muted">{text.noLegality}</p>
{:else}
	<dl class="grid gap-2 text-base">
		{#each entries as entry (entry.label)}
			<dt class="text-text-muted">{entry.label}</dt>
			<dd class="font-mono text-text">{entry.value}</dd>
		{/each}
	</dl>
	<p class="text-sm text-text-muted">{text.declaredNote}</p>
{/if}
