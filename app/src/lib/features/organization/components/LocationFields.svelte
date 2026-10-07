<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { id } from '#lib/i18n/id.js';
	import type { RegionChoice } from '../regions.svelte';
	import type { GeoPoint, OnboardingErrors, PlaceValues } from '../schema';
	import LocationPin from './LocationPin.svelte';

	interface Props {
		regions: RegionChoice;
		values: PlaceValues;
		errors: OnboardingErrors;
		isConsignor: boolean;
	}

	let { regions, values, errors, isConsignor }: Props = $props();

	const text = id.onboarding;
	const placeholder = { value: '', label: text.choose };
	const provinceOptions = $derived([
		placeholder,
		...regions.provinces.map((province) => ({ value: province.code, label: province.name }))
	]);
	const cityOptions = $derived([
		placeholder,
		...regions.cities.map((city) => ({ value: city.code, label: city.name }))
	]);
	const mapCenter = $derived(values.location ?? regions.centerOf(values.cityCode));

	function chooseProvince(provinceCode: string) {
		values.cityCode = '';
		values.location = null;
		void regions.loadCities(provinceCode);
	}

	async function retry() {
		await regions.loadProvinces();
		await regions.loadCities(values.provinceCode);
	}
</script>

{#if regions.failed}
	<Alert title={text.regionsFailed.title}>
		{text.regionsFailed.body}
		{#snippet action()}
			<Button variant="secondary" onclick={() => void retry()}>{id.auth.retry}</Button>
		{/snippet}
	</Alert>
{/if}
<Select
	label={text.province}
	name="province_code"
	options={provinceOptions}
	bind:value={values.provinceCode}
	onchange={chooseProvince}
	error={errors.provinceCode}
/>
<Select
	label={text.city}
	name="city_code"
	options={cityOptions}
	bind:value={values.cityCode}
	onchange={() => {
		values.location = null;
	}}
	error={errors.cityCode}
/>
{#if mapCenter}
	<LocationPin
		center={mapCenter}
		isPinned={values.location !== null}
		{isConsignor}
		onpin={(point: GeoPoint) => {
			values.location = point;
		}}
	/>
{/if}
<TextField
	label={text.district}
	name="district"
	bind:value={values.district}
	error={errors.district}
/>
<TextField
	label={text.address}
	name="address"
	autocomplete="street-address"
	hint={text.addressHint}
	bind:value={values.address}
	error={errors.address}
/>
