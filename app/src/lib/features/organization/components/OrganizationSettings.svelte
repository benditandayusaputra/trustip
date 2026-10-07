<script lang="ts">
	import { tick } from 'svelte';
	import Alert from '#lib/components/ui/Alert.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import ImageField from '#lib/components/ui/ImageField.svelte';
	import Skeleton from '#lib/components/ui/Skeleton.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Organization } from '#lib/api/organizations.js';
	import { id } from '#lib/i18n/id.js';
	import { ProfileSettings, type SaveOutcome } from '../settings.svelte';
	import LocationFields from './LocationFields.svelte';
	import ConsignorProfileSection from './ConsignorProfileSection.svelte';
	import ListingSection from './ListingSection.svelte';
	import PaymentSection from './PaymentSection.svelte';
	import StoreProfileSection from './StoreProfileSection.svelte';

	interface Props {
		organizationId: string;
		canEdit: boolean;
	}

	let { organizationId, canEdit }: Props = $props();

	const text = id.settings;
	const fields = id.onboarding;
	const settings = new ProfileSettings(apiClient);
	const toaster = getToaster();
	let formElement: HTMLFormElement | undefined = $state();

	$effect(() => {
		void settings.load(organizationId);
	});

	function keep(saved: Organization) {
		settings.organization = saved;
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		report(await settings.save());
	}

	function report(outcome: SaveOutcome) {
		if (outcome.kind === 'saved') {
			toaster.show({ tone: 'success', message: text.saved });
			return;
		}
		if (outcome.kind === 'invalid') {
			void tick().then(() => {
				formElement?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
			});
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
{:else if settings.status === 'loading' || !settings.organization}
	<Skeleton />
{:else}
	{@const organization = settings.organization}
	<ListingSection {organization} {canEdit} onsaved={keep} />
	<form bind:this={formElement} onsubmit={handleSubmit} novalidate class="flex flex-col gap-5">
		<p class="text-sm text-text-muted">
			{text.typeLabel}: {fields.types[organization.type].label}. {fields.typeFixed}
		</p>
		{#if !canEdit}
			<p role="status" class="text-sm text-warning">{text.ownersOnly}</p>
		{/if}
		<fieldset disabled={!canEdit} class="flex flex-col gap-5">
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
			<TextField
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
			<Button type="submit" busy={settings.busy}>{text.save}</Button>
		</fieldset>
	</form>
	{#if organization.type === 'STORE'}
		<StoreProfileSection {organization} {canEdit} onsaved={keep} />
	{:else}
		<ConsignorProfileSection {organization} {canEdit} onsaved={keep} />
		<PaymentSection {organization} {canEdit} onsaved={keep} />
	{/if}
{/if}
