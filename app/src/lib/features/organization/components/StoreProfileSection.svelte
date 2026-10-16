<script lang="ts">
	import { untrack } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import NumberField from '#lib/components/ui/NumberField.svelte';
	import Select from '#lib/components/ui/Select.svelte';
	import Switch from '#lib/components/ui/Switch.svelte';
	import TextArea from '#lib/components/ui/TextArea.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Organization } from '#lib/api/organizations.js';
	import { organization as id } from '#lib/i18n/organization.js';
	import { reportSection } from '../section';
	import { StoreProfileEditor } from '../store-profile.svelte';
	import ChoiceList from './ChoiceList.svelte';
	import StorePhotos from './StorePhotos.svelte';

	interface Props {
		organization: Organization;
		canEdit: boolean;
		onsaved: (organization: Organization) => void;
	}

	let { organization, canEdit, onsaved }: Props = $props();

	const text = id.storeProfile;
	const editor = new StoreProfileEditor(
		apiClient,
		untrack(() => organization.store_profile)
	);
	const toaster = getToaster();
	const maxPercent = 50;
	const storeTypeOptions = [
		{ value: '', label: id.onboarding.choose },
		...Object.entries(text.storeTypes).map(([value, label]) => ({ value, label }))
	];

	$effect(() => {
		void editor.loadCategories();
	});

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		reportSection(await editor.save(organization.id), {
			toaster,
			savedMessage: text.saved,
			onsaved
		});
	}
</script>

<form onsubmit={handleSubmit} novalidate class="surface-card flex flex-col gap-5 p-5 compact:p-6">
	<h2 class="text-lg font-semibold text-text">{text.title}</h2>
	<fieldset disabled={!canEdit} class="flex flex-col gap-5">
		<Select
			label={text.storeType}
			name="store_type"
			options={storeTypeOptions}
			bind:value={editor.values.storeType}
			error={editor.errors.storeType}
		/>
		<ChoiceList
			legend={text.categories}
			options={editor.categories.map((category) => ({ value: category.id, label: category.name }))}
			bind:selected={editor.values.categoryIds}
			error={editor.errors.categoryIds}
		/>
		<Switch label={text.acceptsFixedDeposit} bind:checked={editor.values.acceptsFixedDeposit} />
		<Switch label={text.acceptsCommission} bind:checked={editor.values.acceptsCommission} />
		{#if editor.values.acceptsCommission}
			<NumberField
				label={text.commissionMin}
				max={maxPercent}
				bind:value={editor.values.commissionMinPercent}
			/>
			<NumberField
				label={text.commissionMax}
				max={maxPercent}
				bind:value={editor.values.commissionMaxPercent}
				{...editor.errors.commissionMax && { error: editor.errors.commissionMax }}
			/>
		{/if}
		<h3 class="text-base font-semibold text-text">{text.requirements}</h3>
		<Switch
			label={text.requiresDistributionPermit}
			bind:checked={editor.values.requiresDistributionPermit}
		/>
		<Switch label={text.requiresHalal} bind:checked={editor.values.requiresHalal} />
		<Switch label={text.requiresSample} bind:checked={editor.values.requiresSample} />
		<TextArea
			label={text.otherRequirements}
			name="other_requirements"
			bind:value={editor.values.otherRequirements}
			error={editor.errors.text}
		/>
		<TextField
			label={text.openingHours}
			name="opening_hours"
			hint={text.openingHoursHint}
			bind:value={editor.values.openingHours}
		/>
		<StorePhotos
			photos={editor.values.photos}
			busy={editor.isUploading}
			error={editor.errors.photos}
			onadd={(file: File) => void editor.addPhoto(file)}
			onremove={(photoId: string) => {
				editor.removePhoto(photoId);
			}}
		/>
		<Button type="submit" busy={editor.busy}>{text.save}</Button>
	</fieldset>
</form>
