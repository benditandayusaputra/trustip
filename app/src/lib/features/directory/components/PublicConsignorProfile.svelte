<script lang="ts">
	import Card from '#lib/components/ui/Card.svelte';
	import ShareButton from '#lib/components/ui/ShareButton.svelte';
	import type { PublicConsignor } from '#lib/api/public.js';
	import { formatBusinessDate } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import ContactActions from './ContactActions.svelte';
	import DeclaredLegality from './DeclaredLegality.svelte';
	import ProfileHeader from './ProfileHeader.svelte';
	import PublicProductList from './PublicProductList.svelte';
	import LazyReportListing from './LazyReportListing.svelte';

	interface Props {
		consignor: PublicConsignor;
		shareUrl: string;
	}

	let { consignor, shareUrl }: Props = $props();

	const text = id.publicProfile;
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
	<ShareButton title={consignor.name} url={shareUrl} />
	<Card title={text.products}>
		<PublicProductList products={consignor.products} consignorSlug={consignor.slug} />
	</Card>
	<Card title={text.legality}>
		<DeclaredLegality legality={consignor.legality} />
	</Card>
	<LazyReportListing
		target={{ target_type: 'ORGANIZATION', target_id: consignor.id }}
		name={consignor.name}
	/>
</article>
