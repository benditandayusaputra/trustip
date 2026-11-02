<script lang="ts">
	import ChefHat from '@lucide/svelte/icons/chef-hat';
	import Card from '#lib/components/ui/Card.svelte';
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import ShareButton from '#lib/components/ui/ShareButton.svelte';
	import type { PublicConsignor } from '#lib/api/public.js';
	import { formatBusinessDate } from '#lib/i18n/format.js';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { publicProfile as publicProfileText } from '#lib/i18n/public-profile.js';
	import ContactActions from './ContactActions.svelte';
	import DeclaredLegality from './DeclaredLegality.svelte';
	import ProfileHeader from './ProfileHeader.svelte';
	import PublicProductList from './PublicProductList.svelte';
	import LazyReportListing from './LazyReportListing.svelte';

	const id = { ...directoryText, ...publicProfileText };

	interface Props {
		consignor: PublicConsignor;
		shareUrl: string;
	}

	let { consignor, shareUrl }: Props = $props();

	const text = id.publicProfile;
</script>

<article class="flex flex-col gap-6">
	<PageHeading icon={ChefHat} backHref="/cari/penitip?tab=penitip">{consignor.name}</PageHeading>
	<ProfileHeader
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
