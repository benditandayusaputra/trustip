<script lang="ts">
	import { onMount, tick } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import CaptchaField from '#lib/components/ui/CaptchaField.svelte';
	import Checkbox from '#lib/components/ui/Checkbox.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import { watchOnline } from '#lib/offline/network.js';
	import { id } from '#lib/i18n/id.js';
	import { RegisterForm, type SubmitOutcome } from '../register.svelte';

	interface Props {
		siteKey: string;
		onregistered: () => void;
	}

	let { siteKey, onregistered }: Props = $props();

	const text = id.auth.register;
	const form = new RegisterForm(apiClient);
	const toaster = getToaster();
	let formElement: HTMLFormElement;
	let captcha: { reset: () => void } | undefined = $state();
	let isOnline = $state(true);

	onMount(() => watchOnline((online) => (isOnline = online)));

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const outcome = await form.submit();
		if (outcome.kind === 'created') {
			toaster.show({ tone: 'success', message: text.created });
			onregistered();
			return;
		}
		captcha?.reset();
		await report(outcome);
	}

	async function report(outcome: Exclude<SubmitOutcome, { kind: 'created' }>) {
		if (outcome.kind === 'invalid') {
			await tick();
			formElement.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
			return;
		}
		const retry = {
			label: id.auth.retry,
			onclick: () => {
				formElement.requestSubmit();
			}
		};
		toaster.show({
			tone: 'danger',
			message: outcome.message,
			...(outcome.canRetry && { action: retry })
		});
	}
</script>

<form bind:this={formElement} onsubmit={handleSubmit} novalidate class="flex flex-col gap-5">
	<TextField
		label={text.fullName}
		name="full_name"
		autocomplete="name"
		bind:value={form.values.fullName}
		error={form.errors.fullName}
	/>
	<TextField
		label={text.email}
		name="email"
		type="email"
		autocomplete="email"
		bind:value={form.values.email}
		error={form.errors.email}
	/>
	<TextField
		label={text.password}
		name="password"
		type="password"
		autocomplete="new-password"
		hint={text.passwordHint}
		bind:value={form.values.password}
		error={form.errors.password}
	/>
	<TextField
		label={text.whatsapp}
		name="whatsapp"
		type="tel"
		autocomplete="tel"
		hint={text.whatsappHint}
		bind:value={form.values.whatsapp}
		error={form.errors.whatsapp}
	/>
	<Checkbox
		name="privacy_accepted"
		bind:checked={form.values.privacyAccepted}
		error={form.errors.privacyAccepted}
	>
		{text.consentPrefix}
		<a href="/privasi" target="_blank" rel="noopener noreferrer" class="text-accent underline">
			{id.nav.privacy}<span class="sr-only"> {id.auth.opensInNewTab}</span>
		</a>
		{text.consentJoin}
		<a href="/ketentuan" target="_blank" rel="noopener noreferrer" class="text-accent underline">
			{id.nav.terms}<span class="sr-only"> {id.auth.opensInNewTab}</span>
		</a>
	</Checkbox>
	<CaptchaField
		bind:this={captcha}
		{siteKey}
		action="signup"
		error={form.errors.captcha}
		onchange={(token: string) => (form.captchaToken = token)}
	/>
	{#if !isOnline}
		<p role="status" class="text-sm text-warning">{text.offline}</p>
	{/if}
	<Button type="submit" busy={form.busy} disabled={!isOnline}>{text.submit}</Button>
</form>
