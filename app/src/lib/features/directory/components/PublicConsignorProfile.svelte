<script lang="ts">
	import Card from '#lib/components/ui/Card.svelte';
	import type { PublicConsignor } from '#lib/api/public.js';
	import { formatBusinessDate } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import ContactActions from './ContactActions.svelte';
	import ProfileHeader from './ProfileHeader.svelte';
	import PublicProductList from './PublicProductList.svelte';
	import ReportListing from './ReportListing.svelte';

	interface Props {
		consignor: PublicConsignor;
	}

	let { consignor }: Props = $props();

	const text = id.publicProfile;
	const legality = $derived(
		[
			{ label: text.businessId, value: consignor.legality.business_id_no },
			{ label: text.distributionPermit, value: consignor.legality.distribution_permit_no },
			{ label: text.halalCert, value: consignor.legality.halal_cert_no }
		].filter((entry) => entry.value !== null)
	);
</script>

<article class="flex flex-col gap-6">
	<ProfileHeader
		name={consignor.name}
		kind={text.consignorKind}
		place={text.place(consignor.district, consignor.city_name)}
		logoUrl={consignor.logo_url}
	/>
	{#if consignor.description}
		<p class="text-base whitespace-pre-line text-text">{consignor.description}</p>
	{/if}
	<p class="text-sm text-text-muted">
		{text.storePartners(consignor.active_partners)} ·
		{id.directory.joined(formatBusinessDate(consignor.joined_on))}
	</p>
	<ContactActions organizationId={consignor.id} name={consignor.name} canPropose={true} />
	<Card title={text.products}>
		<PublicProductList products={consignor.products} />
	</Card>
	<Card title={text.legality}>
		{#if legality.length === 0}
			<p class="text-base text-text-muted">{text.noLegality}</p>
		{:else}
			<dl class="grid gap-2 text-base">
				{#each legality as entry (entry.label)}
					<dt class="text-text-muted">{entry.label}</dt>
					<dd class="font-mono text-text">{entry.value}</dd>
				{/each}
			</dl>
			<p class="text-sm text-text-muted">{text.declaredNote}</p>
		{/if}
	</Card>
	<ReportListing
		target={{ target_type: 'ORGANIZATION', target_id: consignor.id }}
		name={consignor.name}
	/>
</article>
