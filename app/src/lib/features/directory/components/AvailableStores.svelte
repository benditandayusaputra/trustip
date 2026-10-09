<script lang="ts">
	import MapPin from '@lucide/svelte/icons/map-pin';
	import Button from '#lib/components/ui/Button.svelte';
	import { apiClient } from '#lib/api/client.js';
	import type { PublicProduct } from '#lib/api/public.js';
	import { formatDistance } from '#lib/i18n/format.js';
	import { id } from '#lib/i18n/id.js';
	import { NearbyStores, type GeoPosition } from '../availability.svelte';

	interface Props {
		product: PublicProduct;
		consignorSlug: string;
		locate: () => Promise<GeoPosition>;
	}

	let { product, consignorSlug, locate }: Props = $props();

	const text = id.publicProfile;
	const nearby = new NearbyStores(() => product.stores, {
		client: apiClient,
		locate: () => locate()
	});
	const mapZoom = 18;

	function mapHref(location: GeoPosition): string {
		const { latitude, longitude } = location;
		const marker = new URLSearchParams({ mlat: String(latitude), mlon: String(longitude) });
		return `https://www.openstreetmap.org/?${marker.toString()}#map=${String(mapZoom)}/${String(latitude)}/${String(longitude)}`;
	}
</script>

<div class="flex flex-col gap-4">
	{#if nearby.stores.length === 0}
		<p class="text-base text-text-muted">{text.notAvailable}</p>
	{:else}
		<Button
			variant="secondary"
			busy={nearby.status === 'locating'}
			onclick={() => void nearby.sortFrom(consignorSlug, product.sku)}
		>
			{text.sortNearest}
		</Button>
		{#if nearby.status === 'denied'}
			<p role="status" class="text-sm text-warning">{text.locationDenied}</p>
		{:else if nearby.status === 'failed'}
			<p role="status" class="text-sm text-danger">{text.sortFailed}</p>
		{/if}
		<ul class="flex flex-col gap-4">
			{#each nearby.stores as store (store.name + store.address)}
				<li class="flex flex-col gap-1 border-b border-border pb-4">
					{#if store.profile_slug}
						<a href="/t/{store.profile_slug}" class="text-base font-semibold text-accent"
							>{store.name}</a
						>
					{:else}
						<span class="text-base font-semibold text-text">{store.name}</span>
					{/if}
					<span class="text-sm text-text-muted">
						{[store.address, store.district, store.city_name].filter(Boolean).join(', ')}
						{#if store.distance_m !== null}
							· <span class="font-mono">{formatDistance(store.distance_m)}</span>
						{/if}
					</span>
					{#if store.location}
						<a
							href={mapHref(store.location)}
							target="_blank"
							rel="noopener noreferrer"
							class="inline-flex min-h-11 items-center gap-2 self-start text-sm font-semibold text-accent"
						>
							<MapPin class="size-5" aria-hidden="true" />
							{text.openMap(store.name)}
						</a>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>
