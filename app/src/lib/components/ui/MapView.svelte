<script lang="ts">
	import 'leaflet/dist/leaflet.css';
	import MapPin from '@lucide/svelte/icons/map-pin';
	import type { Map as LeafletMap, Marker, LayerGroup } from 'leaflet';
	import { onMount } from 'svelte';
	import { common as id } from '#lib/i18n/common.js';
	import { markerGroups, type MapMarker, type MarkerGroup } from './map-points';

	interface GeoPoint {
		latitude: number;
		longitude: number;
	}
	interface Props {
		center: GeoPoint;
		label: string;
		unavailableText: string;
		onmove: (center: GeoPoint) => void;
		markers?: MapMarker[];
		selectedId?: string | undefined;
		groupLabel?: (count: number) => string;
		onselect?: (id: string) => void;
	}
	let { center, label, unavailableText, onmove, markers, selectedId, groupLabel, onselect }: Props =
		$props();
	const tileUrl = 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';
	const attribution = '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>';
	const streetZoom = 16;
	const diameterDivisor = 2;
	const maxTileZoom = 19;
	const selectedMarkerPriority = 1000;
	const samePlaceDegrees = 1e-4;
	let container: HTMLDivElement;
	let map = $state<LeafletMap>();
	let leaflet = $state<typeof import('leaflet')>();
	let markerLayer: LayerGroup | undefined;
	let shownMarkers = $state<{ ids: string[]; marker: Marker }[]>([]);
	let isUnavailable = $state(false);

	function createMap(library: typeof import('leaflet')) {
		const created = library.map(container, {
			keyboard: true,
			zoomControl: false,
			scrollWheelZoom: false
		});
		created.setView([center.latitude, center.longitude], streetZoom);
		library.control
			.zoom({ zoomInTitle: id.map.zoomIn, zoomOutTitle: id.map.zoomOut })
			.addTo(created);
		library
			.tileLayer(tileUrl, { attribution, maxZoom: maxTileZoom })
			.on('tileerror', () => {
				isUnavailable = true;
			})
			.addTo(created);
		created.on('moveend', () => {
			if (!markers) reportMove(created);
		});
		markerLayer = library.layerGroup().addTo(created);
		map = created;
		leaflet = library;
	}

	onMount(() => {
		let isMounted = true;
		import('leaflet')
			.then((library) => {
				if (isMounted) createMap(library);
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
		if (isAtCenter(moved)) return;
		const { lat, lng } = moved.getCenter();
		onmove({ latitude: lat, longitude: lng });
	}

	function markerIcon(library: typeof import('leaflet'), count: number) {
		const node = document.createElement('span');
		node.className = 'store-map-marker-number';
		node.textContent = String(count);
		const size = parseFloat(getComputedStyle(container).getPropertyValue('--map-marker-size'));
		return library.divIcon({
			html: node,
			className: 'store-map-marker',
			iconSize: [size, size],
			iconAnchor: [size / diameterDivisor, size]
		});
	}

	function addMarker(library: typeof import('leaflet'), group: MarkerGroup): Marker {
		const point = group.point;
		const marker = library.marker([point.latitude, point.longitude], {
			icon: markerIcon(library, group.members.length),
			title: point.label,
			keyboard: true
		});
		marker.on('click', () => onselect?.(point.id));
		markerLayer?.addLayer(marker);
		const element = marker.getElement();
		if (element) {
			labelMarker(element, point, group.members.length);
			if (groupLabel) element.setAttribute('aria-description', groupLabel(group.members.length));
		}
		return marker;
	}

	function labelMarker(element: HTMLElement, point: MapMarker, count: number) {
		element.setAttribute('aria-label', `${point.label}, ${String(count)}`);
		element.setAttribute('aria-pressed', 'false');
		element.addEventListener('keydown', (event) => {
			if (event.key === ' ') {
				event.preventDefault();
				onselect?.(point.id);
			}
		});
	}

	$effect(() => {
		if (map && !markers && !isAtCenter(map))
			map.setView([center.latitude, center.longitude], streetZoom);
	});

	$effect(() => {
		if (!map || !leaflet || !markers) return;
		const library = leaflet;
		markerLayer?.clearLayers();
		shownMarkers = markerGroups(markers).map((group) => ({
			ids: group.members.map((point) => point.id),
			marker: addMarker(library, group)
		}));
		if (markers.length) {
			const padding = parseFloat(getComputedStyle(container).getPropertyValue('--spacing-8'));
			const bounds = leaflet.latLngBounds(
				markers.map((point) => [point.latitude, point.longitude])
			);
			map.fitBounds(bounds, { padding: [padding, padding], maxZoom: streetZoom });
		}
	});

	$effect(() => {
		for (const { ids, marker } of shownMarkers) {
			const element = marker.getElement();
			marker.setZIndexOffset(
				selectedId !== undefined && ids.includes(selectedId) ? selectedMarkerPriority : 0
			);
			element?.setAttribute(
				'aria-pressed',
				String(selectedId !== undefined && ids.includes(selectedId))
			);
			element?.classList.toggle(
				'store-map-marker-selected',
				selectedId !== undefined && ids.includes(selectedId)
			);
		}
	});
</script>

{#if isUnavailable}<p role="status" class="text-sm text-warning">{unavailableText}</p>{/if}
<div class="relative isolate overflow-hidden rounded-control border border-border-strong">
	<div
		bind:this={container}
		role="application"
		aria-label={label}
		class="w-full bg-surface focus-visible:outline-2 focus-visible:outline-accent {markers
			? 'h-(--directory-map-height)'
			: 'aspect-4/3'}"
	></div>
	{#if !markers}
		<MapPin
			class="pointer-events-none absolute top-1/2 left-1/2 z-500 size-8 -translate-x-1/2 -translate-y-full text-danger"
			aria-hidden="true"
		/>
	{/if}
</div>
