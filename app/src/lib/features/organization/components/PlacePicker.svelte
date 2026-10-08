<script lang="ts">
	import Select from '#lib/components/ui/Select.svelte';
	import { apiClient } from '#lib/api/client.js';
	import { id } from '#lib/i18n/id.js';
	import { locateDevice } from '../geolocation';
	import type { Place, PlaceKind } from '../place';
	import { RegionChoice } from '../regions.svelte';

	interface Props {
		place: Place;
		onchange: (place: Place) => void;
	}

	let { place, onchange }: Props = $props();

	const text = id.placePicker;
	const regions = new RegionChoice(apiClient);
	const coordinateDigits = 5;
	let provinceCode = $state('');
	let locating = $state(false);
	let locateFailed = $state(false);
	const kindOptions = Object.entries(text.kinds).map(([value, label]) => ({ value, label }));
	const provinceOptions = $derived([
		{ value: '', label: text.choose },
		...regions.provinces.map((province) => ({ value: province.code, label: province.name }))
	]);
	const cityOptions = $derived([
		{ value: '', label: text.choose },
		...regions.cities.map((city) => ({ value: city.code, label: city.name }))
	]);

	$effect(() => {
		if (place.kind === 'CITY' && regions.provinces.length === 0) {
			void regions.loadProvinces();
		}
	});

	async function useCurrentPosition() {
		locating = true;
		locateFailed = false;
		try {
			const point = await locateDevice();
			onchange({
				kind: 'HERE',
				lat: point.latitude.toFixed(coordinateDigits),
				lng: point.longitude.toFixed(coordinateDigits),
				cityCode: ''
			});
		} catch {
			locateFailed = true;
		} finally {
			locating = false;
		}
	}

	function chooseKind(kind: string) {
		if (kind === 'HERE') {
			void useCurrentPosition();
			return;
		}
		onchange({ kind: kind as PlaceKind, lat: '', lng: '', cityCode: '' });
	}
</script>

<Select label={text.label} options={kindOptions} value={place.kind} onchange={chooseKind} />
{#if locating}
	<p role="status" class="text-sm text-text-muted">{text.locating}</p>
{:else if locateFailed}
	<p role="status" class="text-sm text-warning">{text.locateFailed}</p>
{/if}
{#if place.kind === 'CITY'}
	<Select
		label={text.province}
		options={provinceOptions}
		value={provinceCode}
		onchange={(code: string) => {
			provinceCode = code;
			void regions.loadCities(code);
		}}
	/>
	<Select
		label={text.city}
		options={cityOptions}
		value={place.cityCode}
		onchange={(cityCode: string) => {
			onchange({ ...place, cityCode });
		}}
	/>
{/if}
