<script lang="ts">
	import 'leaflet/dist/leaflet.css';
	import MapPin from '@lucide/svelte/icons/map-pin';
	import type { Map as LeafletMap } from 'leaflet';
	import { onMount } from 'svelte';

	interface GeoPoint {
		latitude: number;
		longitude: number;
	}

	interface Props {
		center: GeoPoint;
		label: string;
		unavailableText: string;
		onmove: (center: GeoPoint) => void;
	}

	let { center, label, unavailableText, onmove }: Props = $props();

	const tileUrl = 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';
	const attribution = '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>';
	const streetZoom = 16;
	const maxTileZoom = 19;
	const samePlaceDegrees = 1e-4;

	let container: HTMLDivElement;
	let map = $state<LeafletMap>();
	let isUnavailable = $state(false);

	onMount(() => {
		let isMounted = true;
		import('leaflet')
			.then((leaflet) => {
				if (!isMounted) {
					return;
				}
				const created = leaflet.map(container, { keyboard: true, zoomControl: true });
				created.setView([center.latitude, center.longitude], streetZoom);
				leaflet.tileLayer(tileUrl, { attribution, maxZoom: maxTileZoom }).addTo(created);
				created.on('moveend', () => {
					reportMove(created);
				});
				map = created;
			})
			.catch(() => {
				isUnavailable = isMounted;
			});
		return () => {
			isMounted = false;
			map?.remove();
		};
	});

	function isAtCenter(shown: LeafletMap): boolean {
		const { lat, lng } = shown.getCenter();
		return (
			Math.abs(lat - center.latitude) < samePlaceDegrees &&
			Math.abs(lng - center.longitude) < samePlaceDegrees
		);
	}

	function reportMove(moved: LeafletMap) {
		if (isAtCenter(moved)) {
			return;
		}
		const { lat, lng } = moved.getCenter();
		onmove({ latitude: lat, longitude: lng });
	}

	$effect(() => {
		if (map && !isAtCenter(map)) {
			map.setView([center.latitude, center.longitude], streetZoom);
		}
	});
</script>

{#if isUnavailable}
	<p role="status" class="text-sm text-warning">{unavailableText}</p>
{/if}
<div class="relative overflow-hidden rounded-control border border-border-strong">
	<div
		bind:this={container}
		role="application"
		aria-label={label}
		class="aspect-4/3 w-full bg-surface focus-visible:outline-2 focus-visible:outline-accent"
	></div>
	<MapPin
		class="pointer-events-none absolute top-1/2 left-1/2 z-500 size-8 -translate-x-1/2 -translate-y-full text-danger"
		aria-hidden="true"
	/>
</div>
