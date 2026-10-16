<script lang="ts">
	import { untrack } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import ImageField from '#lib/components/ui/ImageField.svelte';
	import TextArea from '#lib/components/ui/TextArea.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import type { Organization } from '#lib/api/organizations.js';
	import { organization as id } from '#lib/i18n/organization.js';
	import { PaymentEditor } from '../payment.svelte';
	import { reportSection } from '../section';

	interface Props {
		organization: Organization;
		canEdit: boolean;
		onsaved: (organization: Organization) => void;
	}

	let { organization, canEdit, onsaved }: Props = $props();

	const text = id.payment;
	const editor = new PaymentEditor(
		apiClient,
		untrack(() => organization.payment)
	);
	const toaster = getToaster();

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
	<div class="flex flex-col gap-1">
		<h2 class="text-lg font-semibold text-text">{text.title}</h2>
		<p class="text-sm text-text-muted">{text.intro}</p>
	</div>
	<fieldset disabled={!canEdit} class="flex flex-col gap-5">
		<TextArea
			label={text.info}
			name="payment_info"
			hint={text.infoHint}
			bind:value={editor.info}
			error={editor.errors.info}
		/>
		<ImageField
			label={text.qr}
			chooseLabel={text.chooseQr}
			imageUrl={editor.qr?.url}
			imageAlt={text.qrAlt}
			hint={id.settings.logoHint}
			error={editor.errors.qr}
			busy={editor.isUploading}
			onselect={(file: File) => void editor.chooseQr(file)}
		/>
		<Button type="submit" busy={editor.busy}>{text.save}</Button>
	</fieldset>
</form>
