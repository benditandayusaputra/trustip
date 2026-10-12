<script lang="ts">
	import MapPin from '@lucide/svelte/icons/map-pin';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import type { Component } from 'svelte';

	interface GeoPoint {
		latitude: number;
		longitude: number;
	}

	interface MapProps {
		center: GeoPoint;
		label: string;
		unavailableText: string;
		onmove: (center: GeoPoint) => void;
	}

	interface Props {
		center: GeoPoint;
		label: string;
		showLabel: string;
		unavailableText: string;
	}

	let { center, label, showLabel, unavailableText }: Props = $props();

	let MapView = $state<Component<MapProps>>();
	let loading = $state(false);
	let failed = $state(false);

	async function load() {
		loading = true;
		failed = false;
		try {
			MapView = (await import('#lib/components/ui/MapView.svelte')).default;
		} catch {
			failed = true;
		} finally {
			loading = false;
		}
	}
</script>

{#if MapView}
	<MapView {center} {label} {unavailableText} onmove={() => undefined} />
{:else if loading}
	<Skeleton shape="map" />
{:else}
	{#if failed}<p role="status" class="text-sm text-text-muted">{unavailableText}</p>{/if}
	<button
		type="button"
		class="inline-flex min-h-11 items-center gap-2 self-start font-semibold text-accent"
		onclick={() => void load()}
	>
		<MapPin class="size-5" aria-hidden="true" />
		{showLabel}
	</button>
{/if}
