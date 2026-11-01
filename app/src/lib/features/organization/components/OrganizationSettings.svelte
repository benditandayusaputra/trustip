<script lang="ts">
	import { tick } from 'svelte';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import ImageField from '#lib/components/ui/ImageField.svelte';
	import ContentSkeleton from '#lib/components/ui/ContentSkeleton.svelte';
	import TextArea from '#lib/components/ui/TextArea.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import ThemeSettings from '#lib/components/ui/ThemeSettings.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Organization } from '#lib/api/organizations.js';
	import { common as commonText } from '#lib/i18n/common.js';
	import { organization as organizationText } from '#lib/i18n/organization.js';
	import { ProfileSettings, type SaveOutcome } from '../settings.svelte';
	import LocationFields from './LocationFields.svelte';
	import ConsignorProfileSection from './ConsignorProfileSection.svelte';
	import ListingSection from './ListingSection.svelte';
	import PaymentSection from './PaymentSection.svelte';
	import StoreProfileSection from './StoreProfileSection.svelte';

	const id = { ...commonText, ...organizationText };

	interface Props {
		organizationId: string;
		canEdit: boolean;
		view: 'profile' | 'preferences';
	}

	let { organizationId, canEdit, view }: Props = $props();

	const text = id.settings;
	const fields = id.onboarding;
	const settings = new ProfileSettings(apiClient);
	const toaster = getToaster();
	let formElement: HTMLFormElement | undefined = $state();
	let details: { save: () => Promise<boolean> } | undefined = $state();
	let editing = $state(false);
	let saving = $state(false);
	let revision = $state(0);

	$effect(() => {
		void settings.load(organizationId);
	});

	function keep(saved: Organization) {
		settings.organization = saved;
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			if (details && !(await details.save())) {
				focusInvalid();
				return;
			}
			report(await settings.save());
		} finally {
			saving = false;
		}
	}

	function cancel() {
		editing = false;
		revision += 1;
		void settings.reset();
	}

	function focusInvalid() {
		void tick().then(() => {
			formElement?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
		});
	}

	function report(outcome: SaveOutcome) {
		if (outcome.kind === 'saved') {
			editing = false;
			toaster.show({ tone: 'success', message: text.saved });
			return;
		}
		if (outcome.kind === 'invalid') {
			focusInvalid();
			return;
		}
		toaster.show({ tone: 'danger', message: outcome.message });
	}
</script>

{#if settings.status === 'failed'}
	<Alert title={text.loadFailed}>
		{settings.failure}
		{#snippet action()}
			<Button variant="secondary" onclick={() => void settings.load(organizationId)}>
				{id.auth.retry}
			</Button>
		{/snippet}
	</Alert>
{:else if view === 'preferences'}
	<ThemeSettings />
	{#if settings.organization}
		<ListingSection organization={settings.organization} {canEdit} onsaved={keep} />
		{#if settings.organization.type === 'CONSIGNOR'}
			<PaymentSection organization={settings.organization} {canEdit} onsaved={keep} />
		{/if}
	{:else}
		<ContentSkeleton variant="form" />
	{/if}
{:else if settings.status === 'loading' || !settings.organization}
	<ContentSkeleton variant="form" />
{:else}
	{@const organization = settings.organization}
	<form bind:this={formElement} onsubmit={handleSubmit} novalidate class="flex flex-col gap-6">
		<section class="surface-card flex flex-col gap-5 p-5 compact:p-6">
			<header class="flex flex-wrap items-start justify-between gap-4">
				<div class="flex flex-col gap-2">
					<h2 class="text-lg font-semibold">{text.general}</h2>
					<p class="text-sm text-text-muted">
						{text.typeLabel}:
						<span class="font-semibold text-text">{fields.types[organization.type].label}</span>. {fields.typeFixed}
					</p>
				</div>
				{#if canEdit}
					{#if editing}
						<Button variant="secondary" disabled={saving} onclick={cancel}>{text.cancel}</Button>
					{:else}
						<Button onclick={() => (editing = true)}>{text.edit}</Button>
					{/if}
				{/if}
			</header>
			{#if !canEdit}
				<p role="status" class="text-sm text-warning">{text.ownersOnly}</p>
			{/if}
			<fieldset disabled={!canEdit || !editing} class="flex flex-col gap-5">
				<ImageField
					label={text.logo}
					chooseLabel={text.chooseLogo}
					imageUrl={settings.logo?.url}
					imageAlt={text.logoAlt(organization.name)}
					hint={text.logoHint}
					error={settings.logoError}
					busy={settings.isUploading}
					onselect={(file: File) => void settings.chooseLogo(file)}
				/>
				<TextField
					label={fields.name}
					name="name"
					bind:value={settings.values.name}
					error={settings.errors.name}
				/>
				<TextArea
					label={fields.description}
					name="description"
					hint={fields.descriptionHint}
					bind:value={settings.values.description}
					error={settings.errors.description}
				/>
				<TextField
					label={fields.whatsapp}
					name="whatsapp"
					type="tel"
					bind:value={settings.values.whatsapp}
					error={settings.errors.whatsapp}
				/>
				<LocationFields
					regions={settings.regions}
					values={settings.values}
					errors={settings.errors}
					isConsignor={organization.type === 'CONSIGNOR'}
				/>
			</fieldset>
		</section>
		{#key revision}
			{#if organization.type === 'STORE'}
				<StoreProfileSection
					bind:this={details}
					{organization}
					canEdit={canEdit && editing}
					onsaved={keep}
				/>
			{:else}
				<ConsignorProfileSection
					bind:this={details}
					{organization}
					canEdit={canEdit && editing}
					onsaved={keep}
				/>
			{/if}
		{/key}
		{#if editing}
			<div class="form-footer">
				<Button type="submit" busy={saving}>{text.save}</Button>
			</div>
		{/if}
	</form>
{/if}
