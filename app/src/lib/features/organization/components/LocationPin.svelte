<script lang="ts">
	import LocateFixed from '@lucide/svelte/icons/locate-fixed';
	import Button from '#lib/components/ui/Button.svelte';
	import MapView from '#lib/components/ui/MapView.svelte';
	import { id } from '#lib/i18n/id.js';
	import { locateDevice } from '../geolocation';
	import type { GeoPoint } from '../schema';

	interface Props {
		center: GeoPoint;
		isPinned: boolean;
		isConsignor: boolean;
		onpin: (point: GeoPoint) => void;
	}

	let { center, isPinned, isConsignor, onpin }: Props = $props();

	const text = id.onboarding;
	let isLocating = $state(false);
	let isDenied = $state(false);
	const isWarning = $derived(isDenied && !isPinned);

	async function useDeviceLocation() {
		isLocating = true;
		isDenied = false;
		try {
			onpin(await locateDevice());
		} catch {
			isDenied = true;
		} finally {
			isLocating = false;
		}
	}
</script>

<div class="flex flex-col gap-2">
	<p class="text-sm font-medium text-text">{text.mapLabel}</p>
	<Button variant="secondary" busy={isLocating} onclick={() => void useDeviceLocation()}>
		<LocateFixed class="size-5" aria-hidden="true" />
		{text.useMyLocation}
	</Button>
	<MapView
		{center}
		label={text.mapLabel}
		unavailableText={text.mapUnavailable}
		onmove={(point: GeoPoint) => {
			onpin(point);
		}}
	/>
	<p role="status" class="text-sm {isWarning ? 'text-warning' : 'text-text-muted'}">
		{#if isPinned}
			{text.pinned}
		{:else if isDenied}
			{text.locationDenied}
		{:else}
			{text.mapHint}
		{/if}
	</p>
	{#if isConsignor}
		<p class="text-xs text-text-muted">{text.consignorPrivacy}</p>
	{/if}
</div>
