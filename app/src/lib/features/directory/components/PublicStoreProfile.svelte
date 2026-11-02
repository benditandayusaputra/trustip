<script lang="ts">
	import MapPin from '@lucide/svelte/icons/map-pin';
	import Store from '@lucide/svelte/icons/store';
	import Card from '#lib/components/ui/Card.svelte';
	import PageHeading from '#lib/components/ui/PageHeading.svelte';
	import ShareButton from '#lib/components/ui/ShareButton.svelte';
	import StoreImage from '#lib/components/ui/StoreImage.svelte';
	import Badge from '#lib/components/ui/Badge.svelte';
	import type { PublicStore } from '#lib/api/public.js';
	import { formatBasisPoints, formatBusinessDate, formatDate } from '#lib/i18n/format.js';
	import { directory as directoryText } from '#lib/i18n/directory.js';
	import { organization as organizationText } from '#lib/i18n/organization.js';
	import { publicProfile as publicProfileText } from '#lib/i18n/public-profile.js';
	import ContactActions from './ContactActions.svelte';
	import LazyMap from './LazyMap.svelte';
	import LazyReportListing from './LazyReportListing.svelte';
	import StoreFacts from './StoreFacts.svelte';
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
	const place = $derived(text.place(store.district, store.city_name));
	const cover = $derived(store.photo_urls[0] ?? store.logo_url);
	const morePhotos = $derived(store.photo_urls.slice(1));
	const facts = $derived([
		{
			label: id.directory.onTimeLabel,
			value:
				store.reputation.on_time_bps === null
					? id.directory.newcomer
					: formatBasisPoints(store.reputation.on_time_bps)
		},
		{ label: id.directory.partnersLabel, value: String(store.reputation.active_partners) }
	]);
</script>

<article class="flex flex-col gap-6">
	<PageHeading icon={Store} backHref="/cari/toko">{store.name}</PageHeading>
	<div class="grid items-start gap-6 lg:grid-cols-12 lg:grid-rows-(--profile-rows)">
		<section class="surface-card overflow-hidden lg:col-span-8">
			<StoreImage
				url={cover}
				loading="eager"
				class="aspect-video w-full bg-accent-surface lg:aspect-21/9"
			/>
			<div class="flex flex-col gap-4 p-5 compact:p-6">
				<div class="flex flex-col gap-1">
					<p class="section-label">{kind}</p>
					{#if place}
						<p class="flex items-center gap-1 text-sm text-text-muted">
							<MapPin class="size-4 shrink-0" aria-hidden="true" />{place}
						</p>
					{/if}
				</div>
				{#if store.description}
					<p class="text-base whitespace-pre-line text-text">{store.description}</p>
				{/if}
				<div class="flex flex-col gap-2 border-t border-border pt-4">
					<h2 class="text-sm font-semibold text-text">{text.accepts}</h2>
					<div class="flex flex-wrap gap-2">
						{#each store.accepted_categories as category (category.id)}
							<Badge tone="neutral" label={category.name} />
						{/each}
					</div>
				</div>
			</div>
		</section>
		<aside class="flex min-w-0 flex-col gap-6 lg:col-span-4 lg:row-span-2">
			<Card title={text.contact}>
				{#snippet actions()}
					<ShareButton title={store.name} url={shareUrl} />
				{/snippet}
				<ContactActions organizationId={store.id} name={store.name} canPropose={true} stacked />
			</Card>
			<Card title={text.about}>
				<StoreFacts {facts} />
				<dl class="flex flex-col gap-4 text-base">
					<div class="flex flex-col gap-1">
						<dt class="text-sm text-text-muted">{text.address}</dt>
						<dd class="text-text">{store.address || place}</dd>
					</div>
					{#if store.opening_hours}
						<div class="flex flex-col gap-1">
							<dt class="text-sm text-text-muted">{text.openingHours}</dt>
							<dd class="text-text">{store.opening_hours}</dd>
						</div>
					{/if}
					<div class="flex flex-col gap-1">
						<dt class="text-sm text-text-muted">{id.directory.joinedLabel}</dt>
						<dd class="text-text">{formatBusinessDate(store.joined_on)}</dd>
					</div>
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
		</aside>
		<div class="grid min-w-0 gap-6 md:grid-cols-2 lg:col-span-8">
			{#if morePhotos.length > 0}
				<ul
					class="grid grid-cols-2 gap-4 md:col-span-2 md:grid-cols-3"
					aria-label={text.photos(store.name)}
				>
					{#each morePhotos as url (url)}
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
			<div class="md:col-span-2">
				<LazyReportListing
					target={{ target_type: 'ORGANIZATION', target_id: store.id }}
					name={store.name}
				/>
			</div>
		</div>
	</div>
</article>
