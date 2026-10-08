<script lang="ts">
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { DirectoryConsignorCard } from '#lib/api/directory.js';
	import { id } from '#lib/i18n/id.js';

	interface Props {
		traits: DirectoryConsignorCard['traits'];
	}

	let { traits }: Props = $props();

	const text = id.consignorDirectory;
	const legality = $derived(traits.legality);
	const hasLegality = $derived(legality.has_distribution_permit || legality.has_halal_certificate);
</script>

<div class="flex flex-col gap-2 text-sm text-text-muted">
	<div class="flex flex-wrap gap-2">
		{#if legality.has_distribution_permit}
			<Badge tone="neutral" label={text.permit} />
		{/if}
		{#if legality.has_halal_certificate}
			<Badge tone="neutral" label={text.halal} />
		{/if}
		{#if legality.can_provide_sample}
			<Badge tone="neutral" label={text.sample} />
		{/if}
	</div>
	{#if hasLegality}
		<p class="text-xs">{text.declared}</p>
	{/if}
	{#if traits.delivery_radius_km !== null}
		<p>{text.deliveryRadius(traits.delivery_radius_km)}</p>
	{/if}
	<p>
		{traits.reputation.is_established
			? text.partners(traits.reputation.active_partners)
			: text.newcomer}
	</p>
</div>
