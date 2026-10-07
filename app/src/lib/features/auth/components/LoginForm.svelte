<script lang="ts">
	import { onMount, tick } from 'svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import CaptchaField from '#lib/components/ui/CaptchaField.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { getToaster } from '#lib/components/ui/toast.svelte.js';
	import { apiClient } from '#lib/api/client.js';
	import { watchOnline } from '#lib/offline/network.js';
	import { id } from '#lib/i18n/id.js';
	import { LoginForm, type LoginOutcome } from '../login.svelte';

	interface Props {
		siteKey: string;
		onsignedin: () => void;
	}

	let { siteKey, onsignedin }: Props = $props();

	const text = id.auth.login;
	const form = new LoginForm(apiClient);
	const toaster = getToaster();
	let formElement: HTMLFormElement;
	let formMessage: HTMLParagraphElement | undefined = $state();
	let captcha: { reset: () => void } | undefined = $state();
	let isOnline = $state(true);

	onMount(() => watchOnline((online) => (isOnline = online)));

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const outcome = await form.submit();
		if (outcome.kind === 'signedIn') {
			onsignedin();
			return;
		}
		captcha?.reset();
		await report(outcome);
	}

	async function report(outcome: Exclude<LoginOutcome, { kind: 'signedIn' }>) {
		if (outcome.kind === 'invalid') {
			await tick();
			(formMessage ?? formElement.querySelector<HTMLElement>('[aria-invalid="true"]'))?.focus();
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
	{#if form.errors.form}
		<p bind:this={formMessage} role="alert" tabindex="-1" class="text-base text-danger">
			{form.errors.form}
		</p>
	{/if}
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
		autocomplete="current-password"
		bind:value={form.values.password}
		error={form.errors.password}
	/>
	<CaptchaField
		bind:this={captcha}
		{siteKey}
		action="login"
		error={form.errors.captcha}
		onchange={(token: string) => (form.captchaToken = token)}
	/>
	{#if !isOnline}
		<p role="status" class="text-sm text-warning">{text.offline}</p>
	{/if}
	<Button type="submit" busy={form.busy} disabled={!isOnline}>{text.submit}</Button>
</form>
