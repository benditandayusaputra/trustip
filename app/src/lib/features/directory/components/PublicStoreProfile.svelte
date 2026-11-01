<script lang="ts">
	import Card from '#lib/components/ui/Card.svelte';
	import ShareButton from '#lib/components/ui/ShareButton.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { PublicStore } from '#lib/api/public.js';
	import { formatBasisPoints, formatBusinessDate, formatDate } from '#lib/i18n/format.js';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { organization as organizationText } from '#lib/i18n/organization.js';
	import { publicProfile as publicProfileText } from '#lib/i18n/public-profile.js';
	import ContactActions from './ContactActions.svelte';
	import ProfileHeader from './ProfileHeader.svelte';
	import LazyMap from './LazyMap.svelte';
	import LazyReportListing from './LazyReportListing.svelte';
	import StoreTermsList from './StoreTermsList.svelte';

	const id = { ...directoryText, ...organizationText, ...publicProfileText };

	interface Props {
		store: PublicStore;
		shareUrl: string;
	}

	let { store, shareUrl }: Props = $props();

	const text = id.publicProfile;
	const kind = $derived(
		store.store_type ? id.storeProfile.storeTypes[store.store_type] : text.storeKind
	);
	const reputation = $derived(
		store.reputation.on_time_bps === null
			? id.directory.newcomer
			: id.directory.onTime(formatBasisPoints(store.reputation.on_time_bps))
	);
</script>

<article class="flex flex-col gap-6">
	<ProfileHeader
		name={store.name}
		{kind}
		place={text.place(store.district, store.city_name)}
		logoUrl={store.logo_url}
		isStore
	/>
	{#if store.description}
		<p class="text-base whitespace-pre-line text-text">{store.description}</p>
	{/if}
	<ContactActions organizationId={store.id} name={store.name} canPropose={true} />
	<ShareButton title={store.name} url={shareUrl} />
	<Card title={text.about}>
		<dl class="grid gap-2 text-base">
			<dt class="text-text-muted">{text.address}</dt>
			<dd class="text-text">{store.address || text.place(store.district, store.city_name)}</dd>
			{#if store.opening_hours}
				<dt class="text-text-muted">{text.openingHours}</dt>
				<dd class="text-text">{store.opening_hours}</dd>
			{/if}
			<dt class="text-text-muted">{text.reputation}</dt>
			<dd class="text-text">
				{reputation} · {id.directory.activePartners(store.reputation.active_partners)} ·
				{id.directory.joined(formatBusinessDate(store.joined_on))}
			</dd>
		</dl>
		{#if store.location}
			<LazyMap
				center={store.location}
				label={text.map(store.name)}
				showLabel={text.showMap}
				unavailableText={text.mapUnavailable}
			/>
		{/if}
	</Card>
	{#if store.photo_urls.length > 0}
		<ul class="grid grid-cols-2 gap-4 md:grid-cols-3" aria-label={text.photos(store.name)}>
			{#each store.photo_urls as url (url)}
				<li>
					<img
						src={url}
						alt=""
						width="400"
						height="300"
						loading="lazy"
						decoding="async"
						class="aspect-4/3 w-full rounded-control object-cover"
					/>
				</li>
			{/each}
		</ul>
	{/if}
	<Card title={text.accepts}>
		<div class="flex flex-wrap gap-2">
			{#each store.accepted_categories as category (category.id)}
				<Badge tone="neutral" label={category.name} />
			{/each}
		</div>
	</Card>
	<Card title={text.terms}>
		<StoreTermsList terms={store.terms} otherRequirements={store.other_requirements} />
	</Card>
	<Card title={text.openNeeds}>
		{#if store.open_needs.length === 0}
			<p class="text-base text-text-muted">{text.noOpenNeeds}</p>
		{:else}
			<ul class="flex flex-col gap-2">
				{#each store.open_needs as need (need.id)}
					<li class="flex flex-col">
						<a href="/kebutuhan/{need.id}" class="text-base font-semibold text-accent"
							>{need.title}</a
						>
						<span class="text-sm text-text-muted">
							{need.category} · {text.needUntil(formatDate(need.expires_at))}
						</span>
					</li>
				{/each}
			</ul>
		{/if}
	</Card>
	<LazyReportListing
		target={{ target_type: 'ORGANIZATION', target_id: store.id }}
		name={store.name}
	/>
</article>
