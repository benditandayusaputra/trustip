<script lang="ts">
	import { untrack } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import NumberField from '#lib/components/ui/NumberField.svelte';
	import Switch from '#lib/components/ui/Switch.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Organization } from '#lib/api/organizations.js';
	import { id } from '#lib/i18n/id.js';
	import { ConsignorProfileEditor } from '../consignor-profile.svelte';
	import { reportSection } from '../section';
	import ChoiceList from './ChoiceList.svelte';

	interface Props {
		organization: Organization;
		canEdit: boolean;
		onsaved: (organization: Organization) => void;
	}

	let { organization, canEdit, onsaved }: Props = $props();

	const text = id.consignorProfile;
	const editor = new ConsignorProfileEditor(
		apiClient,
		untrack(() => organization.consignor_profile)
	);
	const toaster = getToaster();
	const maxRadiusKm = 500;
	const dayOptions = Object.entries(text.days).map(([value, label]) => ({ value, label }));
	const storeTypeOptions = Object.entries(id.storeProfile.storeTypes).map(([value, label]) => ({
		value,
		label
	}));

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		reportSection(await editor.save(organization.id), {
			toaster,
			savedMessage: text.saved,
			onsaved
		});
	}
</script>

<form onsubmit={handleSubmit} novalidate class="flex flex-col gap-5">
	<h2 class="text-lg font-semibold text-text">{text.title}</h2>
	<fieldset disabled={!canEdit} class="flex flex-col gap-5">
		<NumberField
			label={text.deliveryRadius}
			max={maxRadiusKm}
			hint={text.deliveryRadiusHint}
			bind:value={editor.values.deliveryRadiusKm}
			{...editor.errors.deliveryRadius && { error: editor.errors.deliveryRadius }}
		/>
		<ChoiceList
			legend={text.deliveryDays}
			options={dayOptions}
			bind:selected={editor.values.deliveryDays}
			error={editor.errors.deliveryDays}
		/>
		<ChoiceList
			legend={text.preferredStoreTypes}
			options={storeTypeOptions}
			bind:selected={editor.values.preferredStoreTypes}
			error={editor.errors.preferredStoreTypes}
		/>
		<h3 class="text-base font-semibold text-text">{text.legalTitle}</h3>
		<TextField
			label={text.businessIdNo}
			name="business_id_no"
			hint={text.declaredHint}
			bind:value={editor.values.businessIdNo}
			error={editor.errors.businessIdNo}
		/>
		<TextField
			label={text.distributionPermitNo}
			name="distribution_permit_no"
			hint={text.declaredHint}
			bind:value={editor.values.distributionPermitNo}
			error={editor.errors.distributionPermitNo}
		/>
		<TextField
			label={text.halalCertNo}
			name="halal_cert_no"
			hint={text.declaredHint}
			bind:value={editor.values.halalCertNo}
			error={editor.errors.halalCertNo}
		/>
		<Switch label={text.canProvideSample} bind:checked={editor.values.canProvideSample} />
		<Button type="submit" busy={editor.busy}>{text.save}</Button>
	</fieldset>
</form>
