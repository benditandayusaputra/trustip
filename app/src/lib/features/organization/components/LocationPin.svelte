<script lang="ts">
	import LocateFixed from '@lucide/svelte/icons/locate-fixed';
	import Button from '#lib/components/ui/Button.svelte';
	import MapView from '#lib/components/ui/MapView.svelte';
	import { id } from '#lib/i18n/id.js';
	import { locateDevice } from '../geolocation';
	import type { GeoPoint } from '../schema';
	import type { OnboardingWizard } from '../wizard.svelte';

	interface Props {
		wizard: OnboardingWizard;
		center: GeoPoint;
	}

	let { wizard, center }: Props = $props();

	const text = id.onboarding;
	let isLocating = $state(false);
	let isDenied = $state(false);
	const isWarning = $derived(isDenied && !wizard.values.location);

	async function useDeviceLocation() {
		isLocating = true;
		isDenied = false;
		try {
			wizard.pinAt(await locateDevice());
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
		onmove={(point: GeoPoint) => {
			wizard.pinAt(point);
		}}
	/>
	<p role="status" class="text-sm {isWarning ? 'text-warning' : 'text-text-muted'}">
		{#if wizard.values.location}
			{text.pinned}
		{:else if isDenied}
			{text.locationDenied}
		{:else}
			{text.mapHint}
		{/if}
	</p>
	{#if wizard.values.type === 'CONSIGNOR'}
		<p class="text-xs text-text-muted">{text.consignorPrivacy}</p>
	{/if}
</div>
