<script lang="ts">
	import TextArea from '#lib/components/ui/TextArea.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { organization as id } from '#lib/i18n/organization.js';
	import type { OnboardingWizard } from '../wizard.svelte';

	interface Props {
		wizard: OnboardingWizard;
	}

	let { wizard }: Props = $props();

	const text = id.onboarding;
</script>

<TextField
	label={text.name}
	name="name"
	autocomplete="organization"
	bind:value={wizard.values.name}
	error={wizard.errors.name}
/>
<TextArea
	label={text.description}
	name="description"
	hint={text.descriptionHint}
	bind:value={wizard.values.description}
	error={wizard.errors.description}
/>
<TextField
	label={text.whatsapp}
	name="whatsapp"
	type="tel"
	autocomplete="tel"
	hint={text.whatsappHint}
	bind:value={wizard.values.whatsapp}
	error={wizard.errors.whatsapp}
/>
