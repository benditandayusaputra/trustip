<script lang="ts">
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { id } from '#lib/i18n/id.js';
	import type { OnboardingWizard } from '../wizard.svelte';

	interface Props {
		wizard: OnboardingWizard;
	}

	let { wizard }: Props = $props();

	const text = id.onboarding;
	const placeholder = { value: '', label: text.choose };
	const provinceOptions = $derived([
		placeholder,
		...wizard.provinces.map((province) => ({ value: province.code, label: province.name }))
	]);
	const cityOptions = $derived([
		placeholder,
		...wizard.cities.map((city) => ({ value: city.code, label: city.name }))
	]);
</script>

{#if wizard.regionsFailed}
	<Alert title={text.regionsFailed.title}>
		{text.regionsFailed.body}
		{#snippet action()}
			<Button variant="secondary" onclick={() => wizard.retryRegions()}>{id.auth.retry}</Button>
		{/snippet}
	</Alert>
{/if}
<Select
	label={text.province}
	name="province_code"
	options={provinceOptions}
	bind:value={wizard.values.provinceCode}
	onchange={(code: string) => void wizard.chooseProvince(code)}
	error={wizard.errors.provinceCode}
/>
<Select
	label={text.city}
	name="city_code"
	options={cityOptions}
	bind:value={wizard.values.cityCode}
	error={wizard.errors.cityCode}
/>
<TextField
	label={text.district}
	name="district"
	bind:value={wizard.values.district}
	error={wizard.errors.district}
/>
<TextField
	label={text.address}
	name="address"
	autocomplete="street-address"
	hint={text.addressHint}
	bind:value={wizard.values.address}
	error={wizard.errors.address}
/>
